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
  "cf_admin_password": "<admin password>",
  "cf_system_domain": "<system-domain>",
  "cf_istio_domain": "istio.<system-domain>"
}
EOF
```

# Create a plan file
```sh
cat << EOF > "${PWD}/plan.json"
{
  "number_of_apps": 10,
  "app_instances": 1,
  "cleanup": false
}
EOF
```

## Running Tests
```sh
CONFIG="$PWD/config.json" PLAN="$PWD/plan.json" scripts/test
```
## Running CATS
```sh
You can run CATS after running scalling tests as you would usually, however be sure to the set the flag cleanup in the plan file to false and manually delete your org.
```
