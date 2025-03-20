# Access Control System PoC

This project demonstrates a Role-Based Access Control (RBAC) system using Open Policy Agent (OPA) through a proxy-based architecture.

The system uses a proxy-based approach where the Policy Enforcement Point (PEP) acts as a reverse proxy, intercepting all requests to enforce access control before they reach the application.

## Key Learnings & Findings

This PoC demonstrates and validates several important aspects of implementing access control using OPA:

### 1. Authorization Architecture

1. Separation of Concerns
   - How to effectively separate authorization from business logic
   - Benefits of centralized policy management
   - Impact on service maintainability and development speed

2. Proxy-based Enforcement
   - Effectiveness of using a reverse proxy for access control
   - Performance implications of the proxy layer
   - Implementation complexity vs. benefits

### 2. OPA Integration

1. Policy Implementation
   - Writing and managing Rego policies
   - Policy testing and validation approaches
   - Policy versioning and deployment strategies

2. Performance Characteristics
   - Policy evaluation latency
   - Impact on response times
   - Scalability considerations

3. Development Experience
   - Learning curve for Rego language
   - Policy debugging and testing tools
   - Developer workflow improvements

### 3. Access Control Features

1. Field-level Access Control
   - Implementation of fine-grained data filtering
   - Performance impact of field filtering
   - Maintainability of field-level policies

2. Role-based Permissions
   - Flexible role definitions
   - Permission inheritance and hierarchy
   - Role assignment and management

### 4. Real-world Applicability

1. Microservices Integration
   - Service independence and loose coupling
   - Policy consistency across services
   - Deployment and operational considerations

2. Practical Challenges
   - Policy distribution and updates
   - Monitoring and troubleshooting
   - Error handling and fallback strategies

3. Production Readiness
   - Required infrastructure components
   - Operational considerations
   - Performance optimization needs

## Get Started

### Prerequisites

- Docker
- Docker Compose
- Make

### Installation and Setup

1. Clone the repository
```bash
git clone git@github.com:bmf-san/poc-opa-access-control-system.git
```

2. Update `/etc/hosts`:
```sh
127.0.0.1 employee.local
127.0.0.1 pdp.local
127.0.0.1 pep.local
127.0.0.1 pip.local
```

3. Start all services using Docker Compose
```bash
make up
```

Additional commands:
```bash
# View logs from all services
make logs

# View logs from a specific service
make log SERVICE=pep

# Stop all services
make down

# Restart all services
make restart

# Access database CLI
make employee-db  # for employee database
make prp-db      # for prp database

# Run tests
make test
```

## Access Control Model Demonstration

These examples demonstrate how clients interact with the employee service through the PEP proxy. All requests go through pep.local, which enforces access control before proxying allowed requests to employee.local:8083.

Each request generates detailed logs showing:
- Request reception and parsing
- Resource and action identification
- Policy evaluation
- Access decision and request forwarding

### RBAC Examples
```bash
# Manager Role: Can view all employee fields
# John Manager (Engineering Manager)
curl -X GET http://employee.local/employees \
  -H "X-User-ID: 11111111-1111-1111-1111-111111111111"
# Response: 200 OK with filtered data:
{
  "employees": [{
    "id": "11111111-1111-1111-1111-111111111111",
    "name": "John Doe",
    "email": "john.doe@example.com",
    "employment_type": "Full-time",
    "employment_type_id": "11111111-1111-1111-1111-111111111111",
    "department_id": "dep1",
    "department_name": "Engineering",
    "position": "Engineer",
    "joined_at": "2023-01-01T00:00:00Z"
  }]
}

# Employee Role: Can view only id and name fields
# Bob Engineer (regular employee)
curl -X GET http://employee.local/employees \
  -H "X-User-ID: 44444444-4444-4444-4444-444444444444"
# Response: 200 OK with filtered data:
{
  "employees": [{
    "id": "11111111-1111-1111-1111-111111111111",
    "name": "John Doe",
    "employment_type": "Full-time",
  }]
}

# Access to non-controlled resource is denied
# All users (including managers) get 403 for departments and other resources
curl -X GET http://employee.local/departments \
  -H "X-User-ID: 11111111-1111-1111-1111-111111111111"
# Response: 403 Forbidden - Access denied

curl -X GET http://employee.local/invalid_resource \
  -H "X-User-ID: 11111111-1111-1111-1111-111111111111"
# Response: 403 Forbidden - Access denied

# Missing User ID: Bad Request
curl -X GET http://employee.local/employees
# Response: 400 Bad Request - Missing X-User-ID header
```

## Documentation

This project's documentation is organized into several sections:

### Architecture and Design

For detailed technical documentation, including:
- Access Control Architecture and Model
- Component Responsibilities
- Access Control Flow Diagrams
- API Specifications
- Data Model
- OPA Integration Analysis
- Operational Design
- Future Considerations

Please refer to our comprehensive [Design Document](docs/design/DESIGN.md).

### Database Documentation

For database schema details and relationships:
- RBAC Tables
- Entity Relationships

See the database documentation in:
- PRP Database: [docs/db/prp](docs/db/prp/README.md)
- Employee Database: [docs/db/employee](docs/db/employee/README.md)

## Development

### Running Tests
```bash
# Run all tests
make test

# Build and start services with changes
make up
```

### Database Documentation
```bash
# Generate database documentation
make gen-dbdocs
```

## Contribution

Issues and Pull Requests are always welcome.

We would be happy to receive your contributions.

Please review the following documents before making a contribution:

- [CODE_OF_CONDUCT](https://github.com/bmf-san/poc-opa-access-control-system/blob/master/.github/CODE_OF_CONDUCT.md)
- [CONTRIBUTING](https://github.com/bmf-san/poc-opa-access-control-system/blob/master/.github/CONTRIBUTING.md)

## References

- [www.openpolicyagent.org - Open Policy Agent](https://www.openpolicyagent.org/)
- [zenn.dev - OPA/Rego入門](https://zenn.dev/mizutani/books/d2f1440cfbba94)
- [kenfdev.hateblo.jp - アプリケーションにおける権限設計の課題](https://kenfdev.hateblo.jp/entry/2020/01/13/115032)

## License

Based on the MIT License.

[LICENSE](https://github.com/bmf-san/poc-opa-access-control-system/blob/master/LICENSE)

## Author

[bmf-san](https://github.com/bmf-san)

- Email: bmf.infomation@gmail.com
- Blog: [bmf-tech.com](http://bmf-tech.com)
- Twitter: [@bmf-san](https://twitter.com/bmf-san)
