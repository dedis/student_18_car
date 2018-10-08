package main

import cli "gopkg.in/urfave/cli.v1"

/*
This holds the cli-commands so the main-file is less cluttered.
*/

func getCommands() cli.Commands {
	tomlORIP := []cli.Flag{
		cli.StringFlag{
			Name:  "toml, t",
			Usage: "give a toml file with the node to add",
		},
		cli.StringFlag{
			Name:  "addr, a",
			Usage: "give an ip:port pair to add to the roster",
		},
	}
	return cli.Commands{
		{
			Name:    "link",
			Aliases: []string{"ln"},
			Usage:   "create and use links with admin privileges",
			Subcommands: cli.Commands{
				{
					Name:      "pin",
					Aliases:   []string{"p"},
					Usage:     "links using a pin written to the log-file on the conode",
					ArgsUsage: "ip:port [PIN]",
					Action:    linkPin,
				},
				{
					Name:      "addfinal",
					Aliases:   []string{"af"},
					Usage:     "adds a final statement to a linked remote node for use by attendees",
					ArgsUsage: "final_statement.toml ip:port",
					Action:    linkAddFinal,
				},
				{
					Name:      "addpublic",
					Aliases:   []string{"ap"},
					Usage:     "adds a public key to a linked remote node for use by attendees",
					ArgsUsage: "public_key ip:port",
					Action:    linkAddPublic,
				},
				{
					Name:    "keypair",
					Aliases: []string{"kp"},
					Usage:   "create a keypair for usage in private/public key",
					Action:  linkPair,
				},
				{
					Name:    "list",
					Aliases: []string{"ls", "l"},
					Usage:   "show a list of all links stored on this client",
					Action:  linkList,
				},
			},
		},

		{
			Name:    "skipchain",
			Aliases: []string{"sc"},
			Usage:   "work with the underlying skipchain",
			Subcommands: []cli.Command{
				{
					Name:      "create",
					Aliases:   []string{"cr", "c"},
					Usage:     "start a new identity",
					ArgsUsage: "group.toml",
					Flags: []cli.Flag{
						cli.IntFlag{
							Name:  "threshold, thr",
							Usage: "the threshold necessary to add a block",
							Value: 2,
						},
						cli.StringFlag{
							Name:  "name, n",
							Usage: "name of this device in the cisc",
						},
						cli.StringFlag{
							Name:  "private, priv",
							Usage: "give the private key to authenticate",
						},
						cli.StringFlag{
							Name:  "token, tok",
							Usage: "give a pop-token to authenticate",
						},
					},
					Action: scCreate,
				},
				{
					Name:      "join",
					Aliases:   []string{"j"},
					Usage:     "propose to join an existing identity by adding this device-key to the skipchain",
					ArgsUsage: "group.toml id [name]",
					Action:    scJoin,
				},
				{
					Name:      "leave",
					Usage:     "leave the skipchain by removing this device from the identity",
					ArgsUsage: "name [skipchain-id]",
					Action:    scLeave,
				},
				{
					Name:    "list",
					Aliases: []string{"ls", "l"},
					Usage:   "show all stored skipchains",
					Action:  scList,
				},
				{
					Name:      "qrcode",
					Aliases:   []string{"qr"},
					Usage:     "print out the qrcode of the identity-skipchain and a node for contact",
					ArgsUsage: "[skipchain-id]",
					Action:    scQrcode,
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "e, explicit",
							Usage: "Display the explicit IP address of the device",
						},
					},
				},
				{
					Name:    "roster",
					Aliases: []string{"r"},
					Usage:   "change the roster for the skipchain",
					Subcommands: cli.Commands{
						cli.Command{
							Name:      "show",
							Aliases:   []string{"s"},
							Usage:     "shows the current roster of the skipchain",
							ArgsUsage: "[skipchain-id]",
							Action:    scRosterShow,
						},
						cli.Command{
							Name:      "set",
							Usage:     "set the current roster of the skipchain",
							ArgsUsage: "group.toml [skipchain-id]",
							Action:    scRosterSet,
						},
						cli.Command{
							Name:      "add",
							Aliases:   []string{"a"},
							Usage:     "adds a node to the current roster of the skipchain",
							ArgsUsage: "[skipchain-id]",
							Action:    scRosterAdd,
							Flags:     tomlORIP,
						},
						cli.Command{
							Name:      "remove",
							Aliases:   []string{"rm"},
							Usage:     "removes a node from the current roster of the skipchain",
							ArgsUsage: "[skipchain-id]",
							Action:    scRosterRemove,
							Flags:     tomlORIP,
						},
					},
				},
			},
		},

		{
			Name:    "data",
			Aliases: []string{"cfg"},
			Usage:   "updating and voting on data",
			Subcommands: []cli.Command{
				{
					Name:      "clear",
					Aliases:   []string{"c"},
					Usage:     "clear the proposition",
					ArgsUsage: "[skipchain-id]",
					Action:    dataClear,
				},
				{
					Name:      "update",
					Aliases:   []string{"upd"},
					Usage:     "fetch the latest data",
					ArgsUsage: "[skipchain-id]",
					Action:    dataUpdate,
				},
				{
					Name:    "list",
					Aliases: []string{"ls", "l"},
					Usage:   "list existing data and proposed",
					Action:  dataList,
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "propose, p",
							Usage: "will also show proposed data",
						},
						cli.BoolFlag{
							Name:  "details, d",
							Usage: "also show the values of the keys",
						},
					},
				},
				{
					Name:      "vote",
					Aliases:   []string{"v"},
					Usage:     "vote on proposed data",
					ArgsUsage: "[skipchain-id]",
					Action:    dataVote,
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "no, n",
							Usage: "refuse vote",
						},
						cli.BoolFlag{
							Name:  "yes, y",
							Usage: "accept vote",
						},
					},
				},
			},
		},

		{
			Name:    "keyvalue",
			Aliases: []string{"kv"},
			Usage:   "storing and retrieving key/value pairs",
			Subcommands: []cli.Command{
				{
					Name:    "list",
					Aliases: []string{"ls", "l"},
					Usage:   "list all values",
					Action:  kvList,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "key",
							Usage: "only prints the value mapped to this key",
						},
					},
				},
				{
					Name:      "value",
					Aliases:   []string{"v"},
					Usage:     "return the value of a key",
					ArgsUsage: "key [skipchain-id]",
					Action:    kvValue,
				},
				{
					Name:      "add",
					Aliases:   []string{"a"},
					Usage:     "add a new key/value pair",
					ArgsUsage: "key value [skipchain-id]",
					Action:    kvAdd,
				},
				{
					Name:      "file",
					Usage:     "add a key/value pair from a file.Key is given in flag, and value is the file in utf-8.",
					ArgsUsage: "csvFile [skipchain-id]",
					Action:    kvAddFile,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "key",
							Usage: "key name where to add the file. Default is the name of the file",
						},
					},
				},
				{
					Name:      "del",
					Aliases:   []string{"d", "rm"},
					Usage:     "delete a value",
					ArgsUsage: "key [skipchain-id]",
					Action:    kvDel,
				},
			},
		},

		{
			Name:  "ssh",
			Usage: "interacting with the ssh-keys stored in the skipchain",
			Subcommands: []cli.Command{
				{
					Name:      "add",
					Aliases:   []string{"a"},
					Usage:     "adds a new entry to the skipchain",
					Action:    sshAdd,
					ArgsUsage: "hostname [skipchain-id]",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "alias, a",
							Usage: "alias to use for that entry",
						},
						cli.StringFlag{
							Name:  "user, u",
							Usage: "user for that connection",
						},
						cli.StringFlag{
							Name:  "port, p",
							Usage: "port for the connection",
						},
						cli.IntFlag{
							Name:  "security, sec",
							Usage: "how many bits for the key-creation",
							Value: 2048,
						},
					},
				},
				{
					Name:      "del",
					Aliases:   []string{"d", "rm"},
					Usage:     "deletes an entry from the skipchain",
					ArgsUsage: "alias_or_host [skipchain-id]",
					Action:    sshDel,
				},
				{
					Name:      "list",
					Aliases:   []string{"ls"},
					Usage:     "shows all entries for this device",
					Action:    sshLs,
					ArgsUsage: "[skipchain-id]",
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "a,all",
							Usage: "show entries for all devices",
						},
					},
				},
				{
					Name:    "rotate",
					Aliases: []string{"r"},
					Usage:   "renews all keys - only active once the vote passed",
					Action:  sshRotate,
				},
				{
					Name:    "sync",
					Aliases: []string{"s"},
					Usage:   "sync ssh-config and blockchain - interactive",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "toblockchain, tob",
							Usage: "force copy of ssh-config-file to blockchain",
						},
						cli.StringFlag{
							Name:  "toconfig, toc",
							Usage: "force copy of blockchain to ssh-config-file",
						},
					},
					Action: sshSync,
				},
			},
		},

		{
			Name:    "follow",
			Aliases: []string{"f"},
			Usage:   "follow skipchains",
			Subcommands: []cli.Command{
				{
					Name:      "add",
					Aliases:   []string{"a"},
					Usage:     "add a new skipchain",
					ArgsUsage: "group ID service-name",
					Action:    followAdd,
				},
				{
					Name:      "del",
					Aliases:   []string{"d", "rm"},
					Usage:     "delete a skipchain",
					ArgsUsage: "ID",
					Action:    followDel,
				},
				{
					Name:    "list",
					Aliases: []string{"ls", "l"},
					Usage:   "list all skipchains and keys",
					Action:  followList,
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "id-only",
							Usage: "only list the skipchain ID",
						},
					},
				},
				{
					Name:    "update",
					Aliases: []string{"upd"},
					Usage:   "update all skipchains",
					Flags: []cli.Flag{
						cli.IntFlag{
							Name:  "poll, p",
							Value: 0,
							Usage: "poll every n seconds",
						},
					},
					Action: followUpdate,
				},
			},
		},

		{
			Name:      "web",
			Usage:     "add a web-site to a skipchain",
			Aliases:   []string{"w"},
			ArgsUsage: "path/page.html [skipchain-id]",
			Action:    kvAddWeb,
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "inline",
					Usage: "inline all images, css and scripts",
				},
			},
		},

		{
			Name:    "cert",
			Aliases: []string{"c"},
			Usage:   "create and use links with admin privileges",
			Subcommands: cli.Commands{
				{
					Name:      "request",
					Aliases:   []string{"q"},
					Usage:     "request a certificate to letsencrypt and store it to the skipchain",
					ArgsUsage: "domain-name cert-dir www-dir",
					Action:    certRequest,
				},
				{
					Name:      "list",
					Aliases:   []string{"l"},
					Usage:     "List the certificate",
					ArgsUsage: "[Skipchain-ID]",
					Action:    certList,
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "verbose, v",
							Usage: "Display the fullchain certificate",
						},
						cli.BoolFlag{
							Name:  "public, p",
							Usage: "Display the public certificate",
						},
						cli.BoolFlag{
							Name:  "chain, c",
							Usage: "Display the chain certificate",
						},
					},
				},
				{
					Name:      "verify",
					Aliases:   []string{"v"},
					Usage:     "verify the certificate against the root certificate",
					ArgsUsage: "cert-key [Skipchain-ID]",
					Action:    certVerify,
				},
				{
					Name:      "renew",
					Aliases:   []string{"u"},
					Usage:     "renew a certificate",
					ArgsUsage: "cert-key [Skipchain-ID]",
					Action:    certRenew,
				},
				{
					Name:      "revoke",
					Aliases:   []string{"k"},
					Usage:     "revoke and delete a certificate",
					ArgsUsage: "certificate key_name [Skipchain-ID]",
					Action:    certRevoke,
				},
				{
					Name:      "retrieve",
					Aliases:   []string{"r"},
					Usage:     "retrieve the certificate of a given key",
					ArgsUsage: "key [Skipchain-ID]",
					Action:    certRetrieve,
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "directory, d",
							Usage: "Give a directory to write the retrieved certificate",
						},
					},
				},
				{
					Name:      "add",
					Aliases:   []string{"a"},
					Usage:     "add a key/cert pair",
					ArgsUsage: "domain path [Skipchain-ID]",
					Action:    certStore,
				},
			},
		},
	}
}
