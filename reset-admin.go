package main

import (
	"fmt"
	"os"

	dbinit "github.com/bctnry/aegis/pkg/aegis/db/init"
	"github.com/bctnry/aegis/routes"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/term"
)

func ResetAdmin(ctx *routes.RouterContext) {
	dbif, err := dbinit.InitializeDatabase(ctx.Config)
	if err != nil {
		fmt.Printf("Failed to connect to database while resetting admin: %s\n", err.Error())
		return
	}
	fmt.Printf("This utility is here for changing the password of the admin user.\n")
	fmt.Printf("Please enter a new password: ")
	s, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Printf("Failed to read password while resetting admin: %s\n", err.Error())
		return
	}
	hashedS, err := bcrypt.GenerateFromPassword(s, bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("Failed to hash password with bcrypt while resetting admin: %s\n", err.Error())
		return
	}
	err = dbif.UpdateUserPassword("admin", string(hashedS))
	if err != nil {
		fmt.Printf("Failed to update password while resetting admin: %s\n", err.Error())
		return
	}
	fmt.Printf("Done.\n")
}
