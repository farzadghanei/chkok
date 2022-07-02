package chkok

import (
	"os/user"
	"strconv"
)


const EX_OK = 0
const EX_DATAERR = 65
const EX_SOFTWARE = 70
const EX_NOINPUT = 72
const EX_IOERR = 74
const EX_TEMPFAIL = 75
const EX_CONFIG = 78


// getUid returns uid of the specified username, if user doesn't exist, but name is numeric, it's assumed the uid
func getUid(name string) (int, error) {
	var uid int
	var uidstr string
	user, err := user.Lookup(name)
	if err == nil {
		uidstr = user.Uid
	} else {
		uidstr = name
	}
	uid, err = strconv.Atoi(uidstr)
	return uid, err
}

// getGid returns gid of the specified group, if gropu doesn't exist, but name is numeric, it's assumed the gid
func getGid(name string) (int, error) {
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