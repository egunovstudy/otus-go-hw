package main

func main() {
	hostUsers := make(map[string][]string)
	for _, user := range users {
		hostUsers[user.Host] = append(hostUsers[user.Host], user.Name)
	}
}
