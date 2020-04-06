### What Is This?

This repository aims to distribute bitcoin<testnet> for free by a web server connected to a spv wallet!!!

### Install:

1. Get it
   ```
   go get github.com/qshuai/coindis
   ```
   
2. Configure your configuration file 

   ```
   cd $GOPATH/src/github.com/qshuai/coindis
   cp app.sample.yaml app.yaml
   vim app.yaml
   ```

3. Run APP

   ```
   cd $GOPATH/src/github.com/qshuai/coindis
   go build
   ./coindis
   ```

   

