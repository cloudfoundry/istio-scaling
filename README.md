# Istio Scaling
This repository is for testing istio scalability on Cloud Foundry 

### Prerequisites
- Working installation of Go
- Valid `$GOPATH`

# Create a config file
```sh
cat << EOF > "${PWD}/config.json"
{
	"cf_admin_user": "admin",
	"cf_admin_password": <admin password>,
	"cf_internal_apps_domain": "apps.internal",
  "cf_system_domain": "<system-domain>",
	"cf_istio_domain": "istio.<system-domain>",
}
EOF
```

# Create a plan file
```sh
cat << EOF > "${PWD}/plan.json"
{
  "number_of_apps": 10,
  "app_instances": 1
}
EOF
```

## Running Tests
```sh
CONFIG="$PWD/config.json" PLAN="$PWD/plan.json" scripts/test
```
