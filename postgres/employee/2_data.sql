-- Insert departments
INSERT INTO departments (id, name, description) VALUES
('11111111-1111-1111-1111-111111111111', 'Engineering', 'Software development and engineering team'),
('22222222-2222-2222-2222-222222222222', 'HR', 'Human Resources department'),
('33333333-3333-3333-3333-333333333333', 'Sales', 'Sales and business development');

-- Insert employment types
INSERT INTO employment_types (id, name, description) VALUES
('11111111-1111-1111-1111-111111111111', 'Full-time', 'Regular full-time employee'),
('22222222-2222-2222-2222-222222222222', 'Part-time', 'Part-time employee'),
('33333333-3333-3333-3333-333333333333', 'Contract', 'Contract-based employee');

-- Insert employees (matching the users in PRP)
INSERT INTO employees (id, name, email, department_id, employment_type_id, position, joined_at) VALUES
-- Engineering department
('11111111-1111-1111-1111-111111111111', 'John Manager', 'john@example.com', '11111111-1111-1111-1111-111111111111', '11111111-1111-1111-1111-111111111111', 'Engineering Manager', '2024-01-01'),  -- Engineering Manager
('44444444-4444-4444-4444-444444444444', 'Bob Engineer', 'bob@example.com', '11111111-1111-1111-1111-111111111111', '11111111-1111-1111-1111-111111111111', 'Software Engineer', '2024-02-01'),     -- Engineer (John's direct report)

-- HR department
('55555555-5555-5555-5555-555555555555', 'Jane HR', 'jane@example.com', '22222222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111111', 'HR Manager', '2024-01-01'),               -- HR Manager
('77777777-7777-7777-7777-777777777777', 'Alice HR', 'alice@example.com', '22222222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111111', 'HR Specialist', '2024-03-01'),         -- HR Staff

-- Sales department
('66666666-6666-6666-6666-666666666666', 'Sarah Sales', 'sarah@example.com', '33333333-3333-3333-3333-333333333333', '11111111-1111-1111-1111-111111111111', 'Sales Manager', '2024-01-01');       -- Sales Manager
