### What Is This?

This repository aims to distribute BCH for free by setuping a web server connected to a bitcoin cash client!!!

### SetUp:

1. Get this repository:

   ```
   go get github.com/qshuai/coindis
   ```

2. Configure your configuration file 

   ```
   cd $GOPATH/src/github.com/qshuai/coindis
   cp conf/app.conf.sample conf/app.conf
   vim conf/app.conf
   ```

3. Run APP

   ```
   cd $GOPATH/src/github.com/qshuai/coindis
   go build
   ./coindis
   ```

   

