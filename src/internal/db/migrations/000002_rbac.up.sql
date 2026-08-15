CREATE TABLE roles (
    code VARCHAR(30) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE permissions (
    code VARCHAR(100) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE role_permissions (
    role_code VARCHAR(30) NOT NULL REFERENCES roles(code) ON DELETE CASCADE,
    permission_code VARCHAR(100) NOT NULL REFERENCES permissions(code) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (role_code, permission_code)
);

-- Seed default roles to satisfy the FK constraint on existing users (which default to 'customer')
INSERT INTO roles (code, name, description) VALUES
('customer', 'Customer', 'Normal user shopping on the system'),
('admin', 'Admin', 'System administrator with full access'),
('staff', 'Staff', 'Store or customer service staff');

-- Seed default permissions
INSERT INTO permissions (code, name, description) VALUES
('brand:create', 'Create Brand', 'Create new brands'),
('brand:update', 'Update Brand', 'Update existing brands'),
('brand:delete', 'Delete Brand', 'Delete brands'),
('brand:read', 'Read Brand', 'Read brand information'),
('product:create', 'Create Product', 'Create new products'),
('product:update', 'Update Product', 'Update existing products'),
('product:delete', 'Delete Product', 'Delete products'),
('product:read', 'Read Product', 'Read product information'),
('category:create', 'Create Category', 'Create new categories'),
('category:update', 'Update Category', 'Update existing categories'),
('category:delete', 'Delete Category', 'Delete categories'),
('category:read', 'Read Category', 'Read category information');

-- Assign permissions to roles
-- Admin gets all permissions
INSERT INTO role_permissions (role_code, permission_code) VALUES
('admin', 'brand:create'),
('admin', 'brand:update'),
('admin', 'brand:delete'),
('admin', 'brand:read'),
('admin', 'product:create'),
('admin', 'product:update'),
('admin', 'product:delete'),
('admin', 'product:read'),
('admin', 'category:create'),
('admin', 'category:update'),
('admin', 'category:delete'),
('admin', 'category:read');

-- Customer gets read-only permissions
INSERT INTO role_permissions (role_code, permission_code) VALUES
('customer', 'brand:read'),
('customer', 'product:read'),
('customer', 'category:read');

-- Add Foreign Key to users table
ALTER TABLE users ADD CONSTRAINT fk_users_role FOREIGN KEY (role) REFERENCES roles(code);
