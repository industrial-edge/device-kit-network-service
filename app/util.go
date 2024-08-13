/*
 * Copyright (c) Siemens 2022
 * Licensed under the MIT license
 * See LICENSE file in the top-level directory
 */

package app

import (
	"errors"
	"log"
	"os"
	"os/user"
	"strconv"
)

// chownSocket private unix domain socket chowner
func chownSocket(address string, userName string, groupName string) error {
	us, err1 := user.Lookup(userName)
	group, err2 := user.LookupGroup(groupName)
	if err1 == nil && err2 == nil {
		uid, _ := strconv.Atoi(us.Uid)
		gid, _ := strconv.Atoi(group.Gid)
		err3 := os.Chmod(address, os.FileMode.Perm(0660))
		err4 := os.Chown(address, uid, gid)
		if err3 != nil || err4 != nil {
			return errors.New("File permissions failed")
		} else {
			log.Println(uid, " : ", gid)
			return nil
		}
	} else {
		return errors.New("File permissions failed")
	}
}
