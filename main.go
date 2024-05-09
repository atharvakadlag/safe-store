package main

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/net/context"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"

	script "github.com/bitfield/script"
	"github.com/leaanthony/clir"
)

const SHARED_FOLDER string = "1AYZs8cjQ67lCnTrlM2dKx0Zq2jmSSqg4"

func upload(filename string, uploadName string) {
	srv, err := drive.NewService(context.Background(), option.WithCredentialsFile("creds.json"))
	if err != nil {
		log.Fatal("Unable to access Drive API:", err)
	}

	file, err := os.Open(filename)
	if err != nil {
		log.Fatalln(err)
	}

	stat, err := file.Stat()
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	res, err := srv.Files.Create(
		&drive.File{
			Parents: []string{SHARED_FOLDER},
			Name:    filename,
		},
	).Media(file, googleapi.ChunkSize(int(stat.Size()))).Do()

	if err != nil {
		log.Fatalln(err)
	}

	fmt.Printf("uploaded " + uploadName + "\n")

	_, _ = srv.Permissions.Create(res.Id, &drive.Permission{
		Role: "reader",
		Type: "anyone",
	}).Do()

}

func readFromFile(file string) ([]byte, error) {
	return script.File(file).Bytes()
}

// func decode(data string) []byte {
// 	decoded, _ := base64.StdEncoding.DecodeString(data)
// 	return decoded
// }

func main() {
	cli := clir.NewCli("safe-store", "A CLI tool for secure management of files on untrusted web servers", "v0.0.1")

	var file string
	var priv string = "priv.key"
	var pub string = "pub.key"

	/// Generate a keypair
	generateCmd := cli.NewSubCommand("generate", "Generate a keypair")

	generateCmd.StringFlag("pubkey", "The public key file", &pub)
	generateCmd.StringFlag("privkey", "The private key file", &priv)

	generateCmd.Action(func() error {
		// Checking for flags
		pk, sk := GenerateKyber()
		script.Echo(string(pk)).WriteFile(pub)
		script.Echo(string(sk)).WriteFile(priv)
		fmt.Println("Generated keypair")
		return nil
	})

	uploadCmd := cli.NewSubCommand("upload", "Encrypt or decrypt data")
	// Flags
	uploadCmd.StringFlag("file", "The file to process", &file)

	uploadCmd.Action(func() error {
		outfile := file + ".crypt"

		pk, sk := GenerateKyber()
		script.Echo(string(pk)).WriteFile(pub)
		script.Echo(string(sk)).WriteFile(priv)

		data_bytes, err := readFromFile(file)
		if err != nil {
			return err
		}
		data := encode(data_bytes)

		pubKey, err := readFromFile(pub)
		if err != nil {
			fmt.Println("Error: Could not read public key")
			return err
		}
		out := Encrypt(pubKey, string(data))

		script.Echo(out).WriteFile(outfile)

		upload(outfile, outfile)
		upload(pub, outfile+".pub")
		return nil
	})

	err := cli.Run()
	if err != nil {
		fmt.Println(err)
	}
}
