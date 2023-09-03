package mogushan

import (
	ns4 "github.com/noxworld-dev/noxscript/ns/v4"
)

func infinite() ns4.Duration {
	return ns4.Seconds(999999)
}

func areAllAlive(arr []ns4.Obj) bool {
	for _, u := range arr {
		if u.CurrentHealth() > 0 {
			return true
		}
	}
	return false
}
