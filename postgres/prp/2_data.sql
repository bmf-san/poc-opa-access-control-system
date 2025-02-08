INSERT INTO tenants (id, name, created_at, updated_at) VALUES
('11111111-1111-1111-1111-111111111111', 'Tenant A', NOW(), NOW()),
('22222222-2222-2222-2222-222222222222', 'Tenant B', NOW(), NOW());

INSERT INTO departments (id, tenant_id, name, created_at, updated_at) VALUES
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '11111111-1111-1111-1111-111111111111', 'HR', NOW(), NOW()),
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '11111111-1111-1111-1111-111111111111', 'Engineering', NOW(), NOW()),
('cccccccc-cccc-cccc-cccc-cccccccccccc', '22222222-2222-2222-2222-222222222222', 'Sales', NOW(), NOW());

INSERT INTO employment_types (id, tenant_id, name, created_at, updated_at) VALUES
('dddddddd-dddd-dddd-dddd-dddddddddddd', '11111111-1111-1111-1111-111111111111', 'full-time', NOW(), NOW()),
('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', '11111111-1111-1111-1111-111111111111', 'part-time', NOW(), NOW()),
('ffffffff-ffff-ffff-ffff-ffffffffffff', '22222222-2222-2222-2222-222222222222', 'contract', NOW(), NOW());

INSERT INTO users (id, tenant_id, department_id, employment_type_id, name, email, created_at, updated_at) VALUES
('11111111-1111-1111-1111-111111111111', '11111111-1111-1111-1111-111111111111', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'dddddddd-dddd-dddd-dddd-dddddddddddd', 'Alice', 'alice@example.com', NOW(), NOW()),
('22222222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111111', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'dddddddd-dddd-dddd-dddd-dddddddddddd', 'Bob', 'bob@example.com', NOW(), NOW()),
('33333333-3333-3333-3333-333333333333', '22222222-2222-2222-2222-222222222222', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'Charlie', 'charlie@example.com', NOW(), NOW());

INSERT INTO resources (id, tenant_id, name, created_at, updated_at) VALUES
('11111111-1111-1111-1111-111111111111', '11111111-1111-1111-1111-111111111111', 'profile', NOW(), NOW()),
('22222222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111111', 'email', NOW(), NOW()),
('33333333-3333-3333-3333-333333333333', '22222222-2222-2222-2222-222222222222', 'salary', NOW(), NOW());

INSERT INTO actions (id, name) VALUES
('11111111-1111-1111-1111-111111111111', 'create'),
('22222222-2222-2222-2222-222222222222', 'read'),
('33333333-3333-3333-3333-333333333333', 'update'),
('44444444-4444-4444-4444-444444444444', 'delete');

INSERT INTO roles (id, tenant_id, name, created_at, updated_at) VALUES
('55555555-5555-5555-5555-555555555555', '11111111-1111-1111-1111-111111111111', 'admin', NOW(), NOW()),
('66666666-6666-6666-6666-666666666666', '11111111-1111-1111-1111-111111111111', 'employee', NOW(), NOW()),
('77777777-7777-7777-7777-777777777777', '22222222-2222-2222-2222-222222222222', 'manager', NOW(), NOW());

INSERT INTO role_permissions (id, role_id, resource_id, action_id) VALUES
('88888888-8888-8888-8888-888888888888', '55555555-5555-5555-5555-555555555555', '11111111-1111-1111-1111-111111111111', '33333333-3333-3333-3333-333333333333'), -- adminがprofileをupdate可能
('99999999-9999-9999-9999-999999999999', '55555555-5555-5555-5555-555555555555', '11111111-1111-1111-1111-111111111111', '44444444-4444-4444-4444-444444444444'), -- adminがemailをdelete可能
('10101010-1010-1010-1010-101010101010', '66666666-6666-6666-6666-666666666666', '11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222222'); -- employeeがprofileをread可能

INSERT INTO user_roles (id, user_id, role_id, tenant_id) VALUES
('11111111-1111-1111-1111-111111111111', '11111111-1111-1111-1111-111111111111', '55555555-5555-5555-5555-555555555555', '11111111-1111-1111-1111-111111111111'), -- user-1はadmin
('22222222-2222-2222-2222-222222222222', '22222222-2222-2222-2222-222222222222', '66666666-6666-6666-6666-666666666666', '11111111-1111-1111-1111-111111111111'); -- user-2はemployee

INSERT INTO attributes (id, tenant_id, name, created_at, updated_at) VALUES
('33333333-3333-3333-3333-333333333333', '11111111-1111-1111-1111-111111111111', 'employment_type', NOW(), NOW()),
('44444444-4444-4444-4444-444444444444', '11111111-1111-1111-1111-111111111111', 'department', NOW(), NOW());

INSERT INTO abac_policies (id, tenant_id, resource_id, action_id, attribute_id, condition, created_at, updated_at) VALUES
('55555555-5555-5555-5555-555555555555', '11111111-1111-1111-1111-111111111111', '11111111-1111-1111-1111-111111111111', '33333333-3333-3333-3333-333333333333', '33333333-3333-3333-3333-333333333333', '{"employment_type": "full-time"}', NOW(), NOW());

INSERT INTO relationships (id, subject_id, object_id, relation, tenant_id) VALUES
('66666666-6666-6666-6666-666666666666', '11111111-1111-1111-1111-111111111111', '11111111-1111-1111-1111-111111111111', 'owner', '11111111-1111-1111-1111-111111111111'), -- user-1がprofileのowner
('77777777-7777-7777-7777-777777777777', '22222222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111111', 'viewer', '11111111-1111-1111-1111-111111111111'); -- user-2はprofileのviewer
