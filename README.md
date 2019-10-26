# Unshort.link
Prevent short link services from tracking you by un shortening the urls for your. Try it on [unshort.link](https://unshort.link)

```
**Design help wanted**

As you may have recognized the current frontend design of unshort.link is super basic and I would like to improve it. 
If you are a designer send me the designs and I will then implement them for you in HTML, CSS should you not feel comfortable with that technologies. 
If you implement it yourself that's even better
```

## Features

- Access short links for you to prevent the short link providers to track you
- Check links against a blacklist to prevent access to a harmful website hidden behind a short link
- Remove known tracking parameters from the urls behind the short links (e.g. the facebook tracking parameter `utm_source`)
- Remove as many url parameters as possible by keeping the same website result. This helps to remove tracking parameters that are so far unknown

## Building

For using up unshort.link on your own server you need a [working golang installation](https://golang.org/doc/install)

### 1) Generating assets

The assets (html, css, js,...) are directly build into the binary for more portability and an easier usage. You need to 
generate that code by entering `go generate ./...` in the main folder of the project. Because of the blacklist being big
this process can take up to 10 minutes

### 2) Building
   
Building the project works with `go build` in the main folder of the project. (Please keep in mind that you need to generate
the assets first)

## Setup

The building process provides you with an all-inclusive binary. Just enter `./unshort.link` in your console and you should
be up and running

### Available configuration flags

- `--url`: Set the url of the server you are running on (this is only required for the frontend) (Default: `http://localhost:8080`)
- `--port`: Port to start the server on (Default: `8080`)
- `--local`: Use the assets (frontend & blacklist) directly from the filesystem instead of the internal binary storage. This helps during the development of the frontend as you do not have to do `go generate ./...` after every change. This should not be used in production. (Default: `false`)

## Contribution

Checkout the open issues if something fits for you and you would like to work on it. 
If you have an feature idea yourself, please open an issue first so we can discuss if it fits into the vision of unshort.link to prevent you from doing unnecessary work.

Do not hesitate to contact me via [unshortlink@simon-frey.eu](mailto:unshortlink@simon-frey.eu) if you have questions how you contribute to the project.

## Support

For any feature request, bug report or setup question **please use the issues functionality of github!** To contact me personally
write an email to [unshortlink@simon-frey.eu](mailto:unshortlink@simon-frey.eu).