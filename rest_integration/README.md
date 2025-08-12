# Basic http function implementation
Before running, make sure to create:
- A dedicated message destination 
- A function with `http_method` and `http_url` string inputs bound to MD created above
- A playbook using function created in previous step
- And run a playbook somehow with adequate parameters
## Usage:
```bash
go run rest_integration/main.go -tokenId="ad25ac8a..." -tokenSecret="74Ispjoy..." --ip="<ip or hostname here>" -insecure=true -destination="<md api name here>"
```

