# websiteproj
A Instagram rip-off written in Go.

*Note: honestly this is the biggest hack I have ever programmed. don't use it*

# Set-up

    go get github.com/lib/pq
    go get github.com/nfnt/resize
    go get golang.org/x/crypto/bcrypt

    mkdir posts/
    mkdir data/

a database called "userstore" needs to have three tables:
- **users:** id, username, password, email, name, description.
- **posts:** id, uid, inreplyto, postdate, replycount.
- **sessions:** sid, uid.

*Note: the db connection can be edited in the main.go file.*
