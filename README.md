![#9963ff](https://placehold.co/800x15/9963ff/9963ff.png)
![#f222ff](https://placehold.co/800x150/161925/f222ff.png?text=ngn&font=raleway)
![#9963ff](https://placehold.co/800x15/9963ff/9963ff.png)

# ngn <sup>Neon Gopher Notifications</sup>

<img src="https://raw.githubusercontent.com/coltwillcox/ngn/master/pictures/badge-0.jpg" width="800">

### Description

ngn contains two parts: one is the code that will flash the Gopher Badge, the other is the code that will run on your computer (as daemon), and will send notifications to the badge (via USB cable). Now notifications can annoy you on this device as well!
The code was tested on Linux, and I have no idea how it will work on MacOS and Windows.

### Features
-   Listens for notifications on "org.freedesktop.Notifications" interface.
-   Does not prevent notifications on host computer.
-   Flashes eyes (LEDs) on incomming notification.
-   Keeps eyes slightly on while there is at least one notification in history.
-   Displays sender application name, date, time, message, and application icon (if any).
-   Keeps history of last 10 notifications.
-   Navigates through history with Left and Right buttons.
-   Clears complete notification history with A key.
-   Clears single notification with B key.

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
notify-send --icon=/home/user/Pictures/user.jpg "Hello world"
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
