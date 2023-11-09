package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

// encode passwd by bcrypt
func createPassBcrypt(passwd string) string {
	// Generate a new bcrypt hash
	bcryptHash, err := bcrypt.GenerateFromPassword([]byte(passwd), bcrypt.DefaultCost)
	if err != nil {
		return err.Error()
	}
	// hash to hex
	hash := hex.EncodeToString(bcryptHash)
	return string(hash)
}

// encode passwd by sha256
func createPassSha256(passwd string) string {
	// Generate a new sha256 hash
	return fmt.Sprintf("%x", sha256.Sum256([]byte(passwd)))
}

// check passwd by bcrypt
func checkPassBcrypt(passwd string, hash string) bool {
	var err error
	bcryptHash, err := hex.DecodeString(hash)
	if err != nil {
		return err == nil
	}
	err = bcrypt.CompareHashAndPassword(bcryptHash, []byte(passwd))
	return err == nil
}

// check passwd by sha256
func checkPassSha256(passwd string, hash string) bool {
	return createPassSha256(passwd) == hash
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [options] <args>\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	options := flag.String("o", "passwd", "options: [passwd|passwd-check]")
	passwdType := flag.String("t", "bcrypt", "password type:[bcrypt|sha256]")
	passwd := flag.String("p", "", "raw password")
	hash := flag.String("h", "", "hash password")

	flag.Parse()

	// check passwd and hash
	if *options == "passwd" && *passwd == "" {
		fmt.Fprintf(os.Stderr, "Please specify password\n")
		os.Exit(1)
	}
	if *options == "passwd-check" && (*hash == "" || *passwd == "") {
		fmt.Fprintf(os.Stderr, "Please specify password\n")
		os.Exit(1)
	}

	switch *options {
	case "passwd":
		switch *passwdType {
		case "bcrypt":
			fmt.Println(createPassBcrypt(*passwd))
		case "sha256":
			fmt.Println(createPassSha256(*passwd))
		default:
			fmt.Println("passwdType error")
		}
	case "passwd-check":
		switch *passwdType {
		case "bcrypt":
			fmt.Println(checkPassBcrypt(*passwd, *hash))
		case "sha256":
			fmt.Println(checkPassSha256(*passwd, *hash))
		default:
			fmt.Println("passwdType error")
		}
	default:
		fmt.Println("options error")
		usage()
	}

}
