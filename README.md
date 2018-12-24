# Istio Scaling
This repository is for testing istio scalability on Cloud Foundry.

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
  "cf_istio_domain": "istio.<system-domain>",
  "cf_org_name": "some-org",
  "cf_space_name": "some-space"
}
EOF
```
`cf_org_name` and `cf_space_name` are optional. If they are not provided, a
random name will be created.

# Create a plan file
```sh
cat << EOF > "${PWD}/plan.json"
{
  "number_of_apps_to_push": 10,
  "number_of_apps_to_curl": 10,
  "passing_threshold": 99.9,
  "app_instances": 1,
  "app_mem_size": "16M",
  "cleanup": false
}
EOF
```
`number_of_apps_to_curl` must be greater than or equal to `number_of_apps_push`.
If greater than, then the test suite is able to curl already-existing apps which
allows for incremental scaling.

`passing_threshold` is the percentage of curls that must succeed for the test
to pass. e.g. 99.9%

## Running Tests
```sh
CONFIG="$PWD/config.json" PLAN="$PWD/plan.json" scripts/test
```
## Running CATS
```sh
You can run CATS after running scalling tests as you would usually, however be sure to the set the flag cleanup in the plan file to false and manually delete your org.
```

## Assets
- `closer-golang.tgz`: This app closes connections which results in a 503 on curl.
- `hello-golang.tgz`: This app responds with 'hello' on curl.
