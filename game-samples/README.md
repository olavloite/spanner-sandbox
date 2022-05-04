# Spanner Gaming Samples

This repository contains sample code for the following use-cases when using Cloud Spanner for the backend:

- Player creation, login, and skin changes

## How to use this

### Setup infrastructure
A terraform file is provided that creates the appropriate resources for these samples.

Resources that are created:
- Spanner instance and database based on user variables in main.tfvars
- (FUTURE) GKE cluster to run the load generators

To set up the infrastructure, do the following:

- Copy `infrastructure/terraform.tfvars.sample` to `infrastructure/terraform.tfvars`
- Modify `infrastructure/terraform.tfvars` for PROJECT and instance configuration
- `terraform apply` from within infrastructure directory

```
cd infrastructure
cp terraform.tfvars.sample terraform.tfvars
vi terraform.tfvars # modify variables

terraform apply
```

### Player profile sample
- Run the profile service

```
cd src/profile-service
go run .
```

- Generate load using the generators/profile/*.py files

```
cd generators/profile
locustfile -f authentication_server.py

# Point webbrowser to http://localhost:8089 to specify load.
# Host url will be: http://localhost:8080
```

### Generator dependencies

The generators are run by Locust.io, which is a Python framework for generating load.

There are several dependencies required to get the generators to work:

- [pyenv](https://github.com/pyenv/pyenv)
- Python 3

#### PyEnv
Pyenv is used to manage multiple versions of Python, and libraries installed through pip separately for different projects.
Follow the PyEnv [installation guide](https://github.com/pyenv/pyenv#installation) to set this up.

Once PyEnv is setup, then you need to install Python 3.6. For instance, to install the latest (at the moment) of `3.6.15`, do this:

```
pyenv install 3.6.15
pyenv global 3.6.15
python -V
# Python 3.6.15
```

#### Python dependencies
Next, install pip3 dependencies:

```
pip3 install -r requirements
```
