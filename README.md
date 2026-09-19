# Setting Up systemd
## Create systemd service
`sudo nano /etc/systemd/system/campbell-radio-satellite-go.service`
```
[Unit]
Description=campbell-radio-satellite-go

[Service]
Type=simple
User=jcampbell/home/jamble-campbell/go/campbell-radio-satellite-go/start.sh

[Install]
WantedBy=multi-user.target
```
`sudo systemctl enable --now campbell-radio-satellite-go`
`sudo journalctl -xfu campbell-radio-satellite-go`

alias campbell-radio-satellite-go-logs="sudo journalctl -xfu campbell-radio-satellite-go"
alias campbell-radio-satellite-go="sudo systemctl restart campbell-radio-satellite-go && campbell-radio-satellite-go-logs"