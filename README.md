Description:
============

`cryptz` makes it easy to securely communicate with peers. Messages are encyrpted using the receivers public key. Only encrypted messages are saved to the persistent store and the server has no ability to decrypt these messages.

Building:
---------

Drop cryptz into `$GOPATH/src` and run `make install`

Dependencies:
-------------

* The server needs to link against some GPG libraries. Specifically, you will need [libgpg-error][gpg-error], [libassuan][assuan] and [libgpgme][gpgme]. All of these can be installed via homebrew. `brew install libgpg-error libassuan gpgme`
* To decrypt messages the server sends, you will need to install [gpgtools][gpgtools]. You can install it via homebrew using `brew cask install gpgtools`
* The above command installs /usr/local/MacGPG2. If you want to run gpg2 from the command line, make sure you include `/usr/local/MacGPG2/bin` in `$PATH`
* The current datastore is `mongodb`. However, this will be changed to `sqlite` soon.

[gpg-error]: https://www.gnupg.org/related_software/libgpg-error/index.html "GnuPG libgpg-error"
[assuan]: https://www.gnupg.org/related_software/libassuan/index.html "GnuPG libassuan"
[gpgme]: https://www.gnupg.org/related_software/gpgme/index.html "GnuPG gpgme"
[gpgtools]: https://gpgtools.org "GnuPG gpgtools"
[mongodb]: https://www.mongodb.com/ "MongoDB"
[sqlite3]: https://www.sqlite.org/ "SQLite"

Motivation:
-----------

* To improve understanding of go
* To learn cgo
* To learn GnuPG & the GPGME API
* To build a go wrapper around GnuPG
* A project for Zillow Hack Week 11 (Extending it for current Hack Week 12)

