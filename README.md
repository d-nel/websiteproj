# websiteproj
*Note: this is a personal project, and honestly it's the biggest hack I have ever programmed. don't use it*

![](https://raw.githubusercontent.com/d-nel/websiteproj/master/example.png)

# Set-up

    go get github.com/lib/pq
    go get github.com/nfnt/resize
    go get golang.org/x/crypto/bcrypt

    export DATABASE_URL='user=Example dbname=example sslmode=disable'
    export RES_PATH='/path/to/resources'
    export PORT='8080'

    cd $RES_PATH

    mkdir posts/
    mkdir data/

The database needs to have four tables:
- **users:** id, username, password, email, name, description, postcount.
- **posts:** id, uid, inreplyto, postdate, replycount.
- **sessions:** sid, uid.
- **blobs:** name, bytes.

*Note: the resources (static, views, data, and posts folders) all need to be in the same folder.*
