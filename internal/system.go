package chkok

import (
	"os/user"
	"strconv"
)

// ExOK exit code for successful run
const ExOK = 0

// ExDataErr exit code for invalid data
const ExDataErr = 65

// ExSoftware generic failure exit code
const ExSoftware = 70

// ExNoInput exit code for when no input provided
const ExNoInput = 72

// ExTempFail exit code when a runtime error
const ExTempFail = 75

// ExConfig exit code when invlaid configurations
const ExConfig = 78

// getUID returns uid of the specified username, if user doesn't exist, but name is numeric, it's assumed the uid
func getUID(name string) (int, error) {
	var uid int
	var uidstr string
	userInfo, err := user.Lookup(name)
	if err == nil {
		uidstr = userInfo.Uid
	} else {
		uidstr = name
	}
	uid, err = strconv.Atoi(uidstr)
	return uid, err
}

// getGID returns gid of the specified group, if group doesn't exist, but name is numeric, it's assumed the gid
func getGID(name string) (int, error) {
	var gid int
	var gidstr string
	group, err := user.LookupGroup(name)
	if err == nil {
		gidstr = group.Gid
	} else {
		gidstr = name
	}
	gid, err = strconv.Atoi(gidstr)
	return gid, err
}
