package main

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"
)
type SSHTerminal struct {
	Session *ssh.Session
	exitMsg string
	stdout  io.Reader
	stdin   io.Writer
	stderr  io.Reader
}

func main() {
	hostKeyCallback, err := knownhosts.New("C:\\Users\\wangw\\.ssh\\known_hosts")
	if err != nil {
		log.Fatal("could not create hostkeycallback function:", err)
	}

	sshConfig := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password("Itcen2531,."),
		},
		HostKeyCallback: hostKeyCallback,
	}

	client, err := ssh.Dial("tcp", "10.12.12.41:22", sshConfig)
	if err != nil {
		fmt.Println(err)
	}
	defer client.Close()

	err = New(client)
	if err != nil {
		fmt.Println(err)
	}
}
func New(client *ssh.Client)error  {
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	s := SSHTerminal{
		Session: session,
	}

	return s.interactiveSession()
}
func (t *SSHTerminal)interactiveSession()error  {
	defer func() {
		if t.exitMsg == ""{
			fmt.Fprintln(os.Stdout, "the connection was closed on the remote side on ", time.Now().Format(time.RFC822))
		}else {
			fmt.Fprintln(os.Stdout, t.exitMsg)
		}
	}()

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.ECHOCTL:       0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	termFD := int(os.Stdin.Fd())
	w, h, _ := terminal.GetSize(termFD)
	termState, _ := terminal.MakeRaw(termFD)
	defer terminal.Restore(termFD, termState)
	err := t.Session.RequestPty("xterm-256color", h, w, modes)

	// 对接 std
	t.Session.Stdout = os.Stdout
	t.Session.Stderr = os.Stderr
	t.Session.Stdin = os.Stdin

	err = t.Session.Shell()
	if err != nil {
		return err
	}
	err = t.Session.Wait()
	if err != nil {
		return err
	}
	return nil
}
//func (t *SSHTerminal) updateTerminalSize() {
//	go func() {
//		// SIGWINCH is sent to the process when the window size of the terminal has
//		// changed.
//		sigwinchCh := make(chan os.Signal, 1)
//		signal.Notify(sigwinchCh, syscall.SIGWINCH)
//
//		fd := int(os.Stdin.Fd())
//		termWidth, termHeight, err := terminal.GetSize(fd)
//		if err != nil {
//			fmt.Println(err)
//		}
//
//		for {
//			select {
//			// The client updated the size of the local PTY. This change needs to occur
//			// on the server side PTY as well.
//			case sigwinch := <-sigwinchCh:
//				if sigwinch == nil {
//					return
//				}
//				currTermWidth, currTermHeight, err := terminal.GetSize(fd)
//
//				// Terminal size has not changed, don't do anything.
//				if currTermHeight == termHeight && currTermWidth == termWidth {
//					continue
//				}
//
//				t.Session.WindowChange(currTermHeight, currTermWidth)
//				if err != nil {
//					fmt.Printf("Unable to send window-change reqest: %s.", err)
//					continue
//				}
//
//				termWidth, termHeight = currTermWidth, currTermHeight
//
//			}
//		}
//	}()
//}


func publicKeyAuthFunc(kPath string) ssh.AuthMethod {
	keyPath, err := homedir.Expand(kPath)
	if err != nil {
		log.Fatal("find key's home dir failed", err)
	}
	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		log.Fatal("ssh key file read failed", err)
	}
	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatal("ssh key signer failed", err)
	}
	return ssh.PublicKeys(signer)
}