### What Is This?

This repository aims to distribute BCH for free by setuping a web server connected to a bitcoin cash client!!!

### SetUp:

1. Install dependencies:

   ```
   go get github.com/bcext/gcash
   go get github.com/astaxie/beego
   ```

2. Clone this repository:

    ```
    git clone github.com/qshuai/coindis $GOPATH/src/coindis
    ```
3. Configure your configuration file 

   ```
   cd $GOPATH/src/github.com/qshuai/coindis
   cp conf/app.conf.sample conf/app.conf
   vim conf/app.conf
   ```

4. Run APP

   ```
   cd $GOPATH/src/github.com/qshuai/coindis
   go build
   ./coindis
   ```

   

