#!/bin/sh

# Setup PDP dependencies
cd cmd/pdp
go mod tidy
cd ../..

# Setup PEP dependencies
cd cmd/pep
go mod tidy
cd ../..

# Setup PIP dependencies
cd cmd/pip
go mod tidy
cd ../..

# Setup Employee dependencies
cd cmd/employee
go mod tidy
cd ../..

# Setup internal dependencies
cd internal
go mod tidy
cd ..

echo "All dependencies have been set up"
