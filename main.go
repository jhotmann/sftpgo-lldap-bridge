package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/sftpgo/sdk"
	"github.com/zekroTJA/timedmap"
)

const (
	SFTPGO_GROUPS_KEY = "sftpgogroups"
)

var (
	cache *timedmap.TimedMap
)

type LoginData struct {
	Username string   `json:"username"`
	User     sdk.User `json:"user"`
	Password string   `json:"password"`
	IP       string   `json:"ip"`
}

func main() {
	cache = timedmap.New(10 * time.Second)

	log.Println("Starting SFTPGo LLDAP Bridge")
	http.HandleFunc("/login", loginHander)

	err := http.ListenAndServe(":8000", nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}

func loginHander(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var data LoginData
	err := decoder.Decode(&data)
	if err != nil {
		log.Println(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Println("Login attempt by user", data.Username, "from", data.IP)

	// Get matching LLDAP user
	ldapToken, err := getLldapToken(data.Username, data.Password)
	if err != nil {
		log.Printf("Error logging in to LLDAP with user '%s': %v\n", data.Username, err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	ldapUser, err := getLldapUser(ldapToken, data.Username)
	if err != nil || ldapUser.User.Id == "" {
		log.Printf("Error getting LLDAP user data for '%s': %v\n", data.Username, err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Check if member of required group
	if config.Lldap.RequiredGroup != "" {
		meetsRequirement := false
		for _, g := range ldapUser.User.Groups {
			if g.DisplayName == config.Lldap.RequiredGroup {
				meetsRequirement = true
				break
			}
		}
		if !meetsRequirement {
			log.Printf("User does not belong to required group '%s'\n", config.Lldap.RequiredGroup)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	// Update SFTPGo groups cache if necessary
	var sftpgoGroups []sdk.Group
	if !cache.Contains(SFTPGO_GROUPS_KEY) {
		log.Printf("Updating SFTPGo groups cache")
		sftpgoGroups, err = getSftpGroups(0)
		if err != nil {
			log.Printf("Error getting LLDAP user data for '%s': %v\n", data.Username, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		cache.Set(SFTPGO_GROUPS_KEY, sftpgoGroups, 5*time.Minute)
	} else {
		sftpgoGroups, _ = cache.GetValue(SFTPGO_GROUPS_KEY).([]sdk.Group)
	}

	// Convert LLDAP groups to match SFTPGo groups
	lldapGroups := lldapGroupsToSftpGroups(ldapUser.User.Groups, sftpgoGroups)
	groupNames := lo.Map(lldapGroups, func(g sdk.GroupMapping, _ int) string { return g.Name })
	log.Printf("%s belongs to %d LLDAP groups that match to SFTPGo: %s", data.Username, len(lldapGroups), strings.Join(groupNames, ", "))

	if data.User.Username != "" { // User already exists
		if groupsMatch(lldapGroups, data.User.Groups) { // No group changes
			log.Println("Groups match, returning 200")
			w.WriteHeader(http.StatusOK)
		} else { // Groups need updating
			log.Println("User exists but groups different")
			data.User.Groups = lldapGroups
			w.Header().Set("Content-Type", "application/json")
			userJson, _ := json.Marshal(data.User)
			w.Write(userJson)
		}
	} else { // User doesn't exist, create SFTPGo user
		log.Println("Creating user in SFTPGo")

		permissions := make(map[string][]string)
		permissions["/"] = []string{"*"}

		newUser := sdk.User{}
		newUser.Status = 1
		newUser.Groups = lldapGroups
		newUser.Permissions = permissions
		newUser.Username = data.Username
		newUser.Password = data.Password

		w.Header().Set("Content-Type", "application/json")
		userJson, _ := json.Marshal(newUser)
		w.Write(userJson)
	}
}

// Converts LLDAP groups to SFTPGo groups via config 'group-mapptings'
//
// If no primary groups are matched and 'sftpgo.default-primary-group' defined,
// then that group will be appended to the list
func lldapGroupsToSftpGroups(lldapGroups []LldapGroup, sftpgoGroups []sdk.Group) (returnGroupMappings []sdk.GroupMapping) {
	for _, g := range lldapGroups {
		groupName := g.DisplayName
		groupType := 2
		for _, m := range config.GroupMappings {
			if m.Lldap == g.DisplayName {
				groupName = m.Sftpgo
				groupType = m.GroupType
				break
			}
		}

		if sftpgoGroupExists(sftpgoGroups, groupName) {
			returnGroupMappings = append(returnGroupMappings, sdk.GroupMapping{Name: groupName, Type: groupType})
		}
	}

	// Add default primary group if defined and no primary group already
	if _, anyMatch := lo.Find(returnGroupMappings, func(g sdk.GroupMapping) bool { return g.Type == 1 }); !anyMatch &&
		config.Sftpgo.DefaultPrimaryGroup != "" && sftpgoGroupExists(sftpgoGroups, config.Sftpgo.DefaultPrimaryGroup) {

		returnGroupMappings = append(returnGroupMappings, sdk.GroupMapping{Name: config.Sftpgo.DefaultPrimaryGroup, Type: 1})
	}

	return returnGroupMappings
}

func sftpgoGroupExists(sftpgoGroups []sdk.Group, name string) bool {
	for _, sg := range sftpgoGroups {
		if name == sg.Name {
			return true
		}
	}
	return false
}

func groupsMatch(lldap []sdk.GroupMapping, sftpgo []sdk.GroupMapping) bool {
	if len(lldap) != len(sftpgo) {
		return false
	}

	for _, g := range lldap {
		if _, anyMatch := lo.Find(sftpgo, func(sftpG sdk.GroupMapping) bool { return sftpG.Name == g.Name }); !anyMatch {
			return false
		}
	}

	return true
}
