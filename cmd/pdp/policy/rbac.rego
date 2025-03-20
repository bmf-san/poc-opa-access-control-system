package policy.rbac

import future.keywords.if

# Default evaluation result
default result = {"allow": false, "allowed_fields": [], "filtered_data": null}

# Main policy evaluation rule
result = response if {
    # Find allowed role with highest privilege
    roles := [role_id |
        role_id := input.user_roles[_].role_id
        has_access_permission(role_id)
    ]
    trace(sprintf("Found roles with access: %v", [roles]))
    count(roles) > 0
    role_id := max(roles)

    # Get allowed fields for the role
    allowed_fields := get_allowed_fields(role_id)
    trace(sprintf("Allowed fields: %v", [allowed_fields]))

    # Filter data if present
    filtered_data := filter_data(allowed_fields)
    trace(sprintf("Filtered data: %v", [filtered_data]))

    response := {
        "allow": true,
        "allowed_fields": allowed_fields,
        "filtered_data": filtered_data
    }
}

# Filter data based on resource type and allowed fields
filter_data(allowed_fields) = filtered if {
    trace(sprintf("Starting filter_data for resource: %s", [input.resource.name]))
    count(allowed_fields) > 0
    input.data != null

    # Get resource type and its items
    resource_type := input.resource.name
    resource_items := object.get(input.data, resource_type, [])
    trace(sprintf("Processing %s items: %v", [resource_type, resource_items]))

    # Return filtered items only if we have data
    count(resource_items) > 0
    filtered_items := [item |
        item := filter_fields(resource_items[_], allowed_fields)
        count(item) > 0
    ]
    count(filtered_items) > 0
    filtered := {resource_type: filtered_items}

    trace(sprintf("Final filtered data: %v", [filtered]))
}

# Default case when no data is present
default filter_data(allowed_fields) = null

# Generic field filtering
filter_fields(object, allowed_fields) = filtered if {
    trace(sprintf("filter_fields input - object: %v, allowed_fields: %v", [object, allowed_fields]))

    # Find valid fields that exist in the object
    valid_fields := [field |
        field := allowed_fields[_]
        object[field] != null
    ]
    trace(sprintf("Valid fields found: %v", [valid_fields]))
    count(valid_fields) > 0

    # Create filtered object with only allowed fields
    filtered := {field: object[field] |
        field := valid_fields[_]
    }
    trace(sprintf("Filtered result: %v", [filtered]))
}

# Default case for filter_fields
default filter_fields(object, allowed_fields) = {}

# Check access permission
has_access_permission(role_id) = result if {
    trace(sprintf("Checking access for role %s", [role_id]))

    # Find matching permission
    matching_perm := [perm |
        perm := input.role_permissions[_]
        perm.role_id == role_id
        perm.action_id == input.action.id
    ][0]

    # Check resource permission
    result := match_resource_permission(matching_perm)
    trace(sprintf("Access check result: %v", [result]))
}

default has_access_permission(role_id) = false

# Check permission for resource
match_resource_permission(perm) = result if {
    resource_permissions := {
        "employees": "11111111-1111-1111-1111-111111111111"
    }

    expected_id := resource_permissions[input.resource.name]
    result := perm.resource_id == expected_id

    trace(sprintf("Checking permission for resource %s: expected=%s, actual=%s, result=%v",
        [input.resource.name, expected_id, perm.resource_id, result]))
}

default match_resource_permission(perm) = false

# Get list of fields user can access (for response)
get_allowed_fields(role_id) = fields if {
    field_permissions := {
        "11111111-1111-1111-1111-111111111111": { # manager role
            "employees": ["id", "name", "email", "department_id", "department_name", "employment_type_id", "employment_type", "position", "joined_at"]
        },
        "22222222-2222-2222-2222-222222222222": { # employee role
            "employees": ["id", "name", "department_name", "employment_type"]
        }
    }

    role_perms := field_permissions[role_id]
    fields := object.get(role_perms, input.resource.name, [])
    trace(sprintf("Getting allowed fields for role %s and resource %s: %v",
        [role_id, input.resource.name, fields]))
}

default get_allowed_fields(role_id) = []
