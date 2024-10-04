# datadog-cli

## Download your data from DataDog for local analysis

DataDog provides rich visualisations of your metrics, logs and infrastructure but sometimes you need to do a more complex analysis which the UI doesn't enable.
`datadog-cli` make it easy to download your data and work on it locally.

# Modules

`datadog-cli` will eventually support downloading all data from DataDog.
Today the following modules are available:
* Metrics: Specify a query to run and download the timeseries for it.
* Containers: Download your current list of running containers, either as a raw list, or grouped by a given key (e.g. `image_name`)

# Building
```
go mod download
go build -o datadog-cli cmd/datadog-cli/
```

# Usage
## API and app keys
Set `DD_API_KEY` and `DD_APP_KEY` in your environment. These are used to authenticate with DataDog.

## Calling modules
Each module has it's own help available.
The syntax for running `datadog-cli` is:
```
./datadog-cli <module> [-<options>...]
```
Example:
```
./datadog-cli metrics -help
```

You can see a list of all available modules by running `datadog-cli` with no module specified.

# Creating modules
Make a copy of `modules/example_module.go`.
The comments in the file should guide you through the steps of customising your module.
Edit `modules/registry.go` and call `RegisterModule()` with your new module's registration method.

Your module has to implement a `Run()` method, which will be its entry point.
You can specify any command-line switches you want your module to expose in the `ParseFalgs()` method.
