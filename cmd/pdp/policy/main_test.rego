package policy

import data.policy.rbac
import future.keywords.if

# Test data for RBAC test cases
data_employees := {
    "employees": [{
        "id": "11111111-1111-1111-1111-111111111111",
        "name": "John Doe",
        "email": "john.doe@example.com",
        "department_id": "dep1",
        "department_name": "Engineering",
        "employment_type_id": "type1",
        "employment_type": "Full-time",
        "position": "Engineer",
        "joined_at": "2023-01-01T00:00:00Z"
    }]
}

# RBAC Test Cases
test_rbac_manager_can_view_all_employee_fields if {
    result := rbac.result with input as {
        "user": {
            "id": "11111111-1111-1111-1111-111111111111"  # John Manager
        },
        "user_roles": [{
            "role_id": "11111111-1111-1111-1111-111111111111"  # manager role
        }],
        "role_permissions": [{
            "role_id": "11111111-1111-1111-1111-111111111111",  # manager role
            "resource_id": "11111111-1111-1111-1111-111111111111",  # employees resource
            "action_id": "11111111-1111-1111-1111-111111111111"  # view action
        }],
        "resource": {
            "id": "11111111-1111-1111-1111-111111111111",  # employees resource
            "name": "employees"
        },
        "action": {
            "id": "11111111-1111-1111-1111-111111111111",  # view action
            "name": "view"
        },
        "data": data_employees
    }

    result.allow
    result.allowed_fields == ["id", "name", "email", "department_id", "department_name", "employment_type_id", "employment_type", "position", "joined_at"]
    result.filtered_data.employees[0].id == data_employees.employees[0].id
    result.filtered_data.employees[0].email == data_employees.employees[0].email
    result.filtered_data.employees[0].employment_type_id == data_employees.employees[0].employment_type_id
}

test_rbac_employee_can_view_limited_employee_fields if {
    result := rbac.result with input as {
        "user": {
            "id": "44444444-4444-4444-4444-444444444444"  # Bob Engineer
        },
        "user_roles": [{
            "role_id": "22222222-2222-2222-2222-222222222222"  # employee role
        }],
        "role_permissions": [{
            "role_id": "22222222-2222-2222-2222-222222222222",  # employee role
            "resource_id": "11111111-1111-1111-1111-111111111111",  # employees resource
            "action_id": "11111111-1111-1111-1111-111111111111"  # view action
        }],
        "resource": {
            "id": "11111111-1111-1111-1111-111111111111",  # employees resource
            "name": "employees"
        },
        "action": {
            "id": "11111111-1111-1111-1111-111111111111",  # view action
            "name": "view"
        },
        "data": data_employees
    }

    result.allow
    result.allowed_fields == ["id", "name", "department_name", "employment_type"]
    result.filtered_data.employees[0].id == data_employees.employees[0].id
    result.filtered_data.employees[0].name == data_employees.employees[0].name
    result.filtered_data.employees[0].department_name == data_employees.employees[0].department_name
    result.filtered_data.employees[0].employment_type == data_employees.employees[0].employment_type
    not result.filtered_data.employees[0].email
    not result.filtered_data.employees[0].department_id
    not result.filtered_data.employees[0].employment_type_id
}

test_rbac_deny_departments_access if {
    result := rbac.result with input as {
        "user": {"id": "44444444-4444-4444-4444-444444444444"}, # Bob Engineer
        "user_roles": [{"role_id": "22222222-2222-2222-2222-222222222222"}], # employee role
        "resource": {
            "id": "22222222-2222-2222-2222-222222222222",  # departments resource
            "name": "departments"
        },
        "action": {
            "id": "11111111-1111-1111-1111-111111111111", # view action
            "name": "view"
        },
        "role_permissions": []
    }

    not result.allow
    result.filtered_data == null
}

test_rbac_deny_invalid_resource if {
    result := rbac.result with input as {
        "user": {"id": "11111111-1111-1111-1111-111111111111"}, # John Manager
        "user_roles": [{"role_id": "11111111-1111-1111-1111-111111111111"}], # manager role
        "resource": {
            "id": "33333333-3333-3333-3333-333333333333", # invalid resource
            "name": "invalid_resource"
        },
        "action": {
            "id": "11111111-1111-1111-1111-111111111111", # view action
            "name": "view"
        },
        "role_permissions": []
    }

    not result.allow
    result.filtered_data == null
}
