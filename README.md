Description:
============

`zecure` makes it easy to securely communicate with peers on the same local network. Messages sent to peers are encyrpted using the targets public key.

Building:
---------

Drop zecure into `$GOPATH/src` and run `make install`

Dependencies:
-------------

* `brew cask install gpgtools`
* The above command installs /usr/local/MacGPG2. If you want to run gpg2 from the command line, make sure you include `/usr/local/MacGPG2/bin` in `$PATH`
