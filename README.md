#### sharedkitchenordersystem

This application is built on golang technologies. The build and launch were done using Docker for easy portability

This project could be built and launched using either:
 - Docker
 - Direct CLI

## Docker
Docker file is crated with golang as base to be run on linux container

step1: run the following command

`docker build --tag sharedkitchendocker <project absolute path>\sharedkitchenordersystem`

If you wish to bypass docker image cache then run this:
`docker build --tag sharedkitchendocker --no-cache <project absolute path>\sharedkitchenordersystem`

for example:
 - with cache: `docker build --tag sharedkitchendocker <project absolute path>\sharedkitchenordersystem`
 - with no cache: with no cache` C:\projects\golang\sharedkitchenordersystem> docker build --tag sharedkitchendocker --no-cache C:\projects\golang\sharedkitchenordersystem`

step2: run the following command to launch the application 
`docker run -e noOfOrdersToRead=10 sharedkitchendocker`

`noOfOrdersToRead` is the configuration to tell the application that how many orders need to be processed at once by the kitchen. Please note: Give a valid positive integer. If an invalid value is given thge default value set in the application  is used 

## Direct CLI

Make sure go is installed in the host machine

step1: build the application using  following command
`go build  .\cmd\sharedkitchenordersystem\main.go`

For example: `C:\projects\golang\sharedkitchenordersystem> go build  .\cmd\sharedkitchenordersystem\main.go`

step 2: run the application with the flags supplied as below

`go run .\cmd\sharedkitchenordersystem\main.go -noOfOrdersToRead=100`

For example: 
`C:\projects\golang\sharedkitchenordersystem> go run .\cmd\sharedkitchenordersystem\main.go -noOfOrdersToRead=100`

Note: The second step performs both build and run, so the first step is just optional but recommended
