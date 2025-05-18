## To run
After downloading the project, cd into the project folder run it with
```
go run main.go
```
The chatroom will be opened on port 8080 on the local IP of your computer. You can check this by running
```
ipconfig
```
On your local machine, the chatroom will be on the IPv4 address. You can enter the chatroom by running
```
telnet {ip address} 8080
```
If telnet is not enabled on your machine, open powershell as an administrator and run the following command:
```
dism /online /Enable-Feature /FeatureName:TelnetClient
```
Then rerun the telnet command and you will enter the LAN chatroom!
