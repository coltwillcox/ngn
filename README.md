![#9963ff](https://placehold.co/800x15/9963ff/9963ff.png)
![#f222ff](https://placehold.co/800x150/161925/f222ff.png?text=ngn&font=raleway)
![#9963ff](https://placehold.co/800x15/9963ff/9963ff.png)

# ngn <sup>Neon Gopher Notifications</sup>

<img src="https://raw.githubusercontent.com/coltwillcox/ngn/master/pictures/badge-0.jpg" width="800">

### Prerequisites

-   Golang: https://go.dev/
-   Tinygo: https://tinygo.org/
-   Gopher Badge (battery not required): https://gopherbadge.com/
-   USB cable

### Run and build process

Clone this repo using:
```shell
git clone git@github.com:coltwillcox/ngn.git
```

Change directories into the ngn directory:
```shell
cd ngn
```

Make sure Gopher Badge is connected, then flash it:
```shell
tinygo flash -size short -target gopher-badge ./gopherbadge/main.go
```

Run daemon:
```shell
go run ./daemon/main.go 
```

Test notifications:
```shell
notify-send "Hello world"
```

Build deamon:
```shell
go build ./daemon/main.go 
```
then add file `main` to startup item.

![#9963ff](https://placehold.co/800x15/9963ff/9963ff.png)

### Daemon <sup>daemon/main.go</sup>

Info:

Screens:

![#9963ff](https://placehold.co/800x15/9963ff/9963ff.png)

### Gopher Badge <sup>gopherbadge/main.go</sup>

Info:

Screens:

![#9963ff](https://placehold.co/800x15/9963ff/9963ff.png)
