Description:
============

[cryptzd][cryptzd] makes it easy to securely communicate with peers. Messages are encyrpted using the receivers public key. Only encrypted messages are saved to the persistent store and the server does not have the  ability to decrypt these messages.

During hackweek 12, I added the `project`, `project_member` and `project_credential` entities. A `project` is just a container for `members` and `credentials`. members are users that have access to the projets credentials. credentials are the entities we're protecting. These could be things like database passwords, API credentials etc...

In order to make it easy to interact with the server, I created a CLI during hackweek 12. That project can be found here [cryptz][cryptz]

[cryptzd]: https://github.com/rajivnavada/cryptzd
[cryptz]: https://github.com/rajivnavada/cryptz

Dependencies:
-------------

* The server needs to link against some GPG libraries. Specifically, you will need [libgpg-error][gpg-error], [libassuan][assuan] and [libgpgme][gpgme]. All of these can be installed via homebrew. `brew install libgpg-error libassuan gpgme`
* To decrypt messages the server sends, you will need to install [gpgtools][gpgtools]. You can install it via homebrew using `brew cask install gpgtools`
* The above command installs /usr/local/MacGPG2. If you want to run gpg2 from the command line, make sure you include `/usr/local/MacGPG2/bin` in `$PATH`
* [sqlite][sqlite] is the datastore.

[gpg-error]: https://www.gnupg.org/related_software/libgpg-error/index.html "GnuPG libgpg-error"
[assuan]: https://www.gnupg.org/related_software/libassuan/index.html "GnuPG libassuan"
[gpgme]: https://www.gnupg.org/related_software/gpgme/index.html "GnuPG gpgme"
[gpgtools]: https://gpgtools.org "GnuPG gpgtools"
[sqlite3]: https://www.sqlite.org/ "SQLite"

Building:
---------

Drop cryptz into `$GOPATH/src` and run `make install`

Alternatively, if you have the GnuPG related dependencies installed, you can just `go get -u github.com/rajivnavada/cryptzd`

Motivation:
-----------

* To improve understanding of go
* To learn cgo
* To learn GnuPG & the GPGME API
* To build a go wrapper around GnuPG
* A project for Zillow Hack Week 11 (Extending it for current Hack Week 12)

Disclaimer:
-----------

This code was written in during 2 hackweeks at Zillow Group. The first was in March 2016 and the second was in July 2016. So parts of the codebase may seem senseless (and it probably is). There are also no tests but hopefully the code is written in a testable manner. I plan to clean up the code and add tests over the next few months as I find the time.

Running:
--------

Once you have `cryptzd` installed, you need to create certificates for the server to use. If you navigate to the src directory of the project, you can call `make cert.pem` to generate the certificates. Then start the project by running `cryptzd -debug`. You can modify the host/port to which the server should bind by using the appropriate flags.

Now that the server is up, you can log into the system by using your ASCII armored GPG public key. The server will try to send an email containing an activation token. The email is encrypted to the key used to log in. If no valid email address or password was provided to the mailer, the content of the email will simply be dumped onto standard output. Decrypt the cipher using the appropriate private key and follow the activation URL. Once you've activated your key, you are ready to use the system. You can now use the [cryptz client][cryptz] to interact with the server.

TIP: `gpg2 --armor --export $KEY_ID | pbcopy` will allow you to copy your public key to the system clipboard on OSX.

License
-------

MIT

