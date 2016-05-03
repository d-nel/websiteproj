# websiteproj
*Note: this is a personal project, and honestly it's the biggest hack I have ever programmed. don't use it*

![](https://raw.githubusercontent.com/d-nel/websiteproj/master/example.png)

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

*const path* in main.go is needed to locate the static, views, data, and posts folders. (currently a work around)
