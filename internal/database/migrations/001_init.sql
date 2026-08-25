-- Initial schema for Useria SaaS OAuth2 multi-tenant platform

CREATE TABLE users (
    id CHAR(36) PRIMARY KEY COMMENT 'UUID',
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) COMMENT 'NULL for OAuth users',
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    avatar_url VARCHAR(255),
    oauth_provider VARCHAR(50) COMMENT 'google, github, or local',
    oauth_id VARCHAR(255),
    status ENUM('active', 'inactive', 'suspended') DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_email (email),
    INDEX idx_oauth_provider (oauth_provider, oauth_id),
    UNIQUE KEY uq_oauth (oauth_provider, oauth_id)
);

CREATE TABLE companies (
    id CHAR(36) PRIMARY KEY COMMENT 'UUID',
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    logo_url VARCHAR(255),
    website VARCHAR(255),
    status ENUM('active', 'inactive', 'suspended') DEFAULT 'active',
    plan VARCHAR(50) DEFAULT 'free' COMMENT 'free, pro, enterprise',
    created_by CHAR(36) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL COMMENT 'soft delete',
    INDEX idx_slug (slug),
    INDEX idx_status (status),
    CONSTRAINT fk_companies_created_by FOREIGN KEY (created_by) REFERENCES users(id)
);

CREATE TABLE roles (
    id CHAR(36) PRIMARY KEY COMMENT 'UUID',
    company_id CHAR(36) COMMENT 'NULL for system roles',
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_system BOOLEAN DEFAULT FALSE COMMENT 'System roles: admin, member, viewer',
    is_custom BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_role_company (company_id, name),
    INDEX idx_company_id (company_id),
    CONSTRAINT fk_roles_company FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE
);

CREATE TABLE company_members (
    id CHAR(36) PRIMARY KEY COMMENT 'UUID',
    company_id CHAR(36) NOT NULL,
    user_id CHAR(36) NOT NULL,
    role_id CHAR(36) NOT NULL,
    status ENUM('active', 'invited', 'inactive') DEFAULT 'active',
    invited_at TIMESTAMP NULL,
    joined_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uq_company_user (company_id, user_id),
    INDEX idx_company_id (company_id),
    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    CONSTRAINT fk_members_company FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE,
    CONSTRAINT fk_members_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_members_role FOREIGN KEY (role_id) REFERENCES roles(id)
);

CREATE TABLE permissions (
    id CHAR(36) PRIMARY KEY COMMENT 'UUID',
    code VARCHAR(100) UNIQUE NOT NULL COMMENT 'users:read, users:write, etc',
    name VARCHAR(100),
    description TEXT,
    category VARCHAR(50) COMMENT 'users, billing, settings, etc',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_code (code),
    INDEX idx_category (category)
);

CREATE TABLE role_permissions (
    id CHAR(36) PRIMARY KEY COMMENT 'UUID',
    role_id CHAR(36) NOT NULL,
    permission_id CHAR(36) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uq_role_permission (role_id, permission_id),
    INDEX idx_role_id (role_id),
    CONSTRAINT fk_role_perms_role FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    CONSTRAINT fk_role_perms_perm FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
);

CREATE TABLE oauth_tokens (
    id CHAR(36) PRIMARY KEY COMMENT 'UUID',
    user_id CHAR(36) NOT NULL,
    company_id CHAR(36) NOT NULL,
    access_token VARCHAR(1000) NOT NULL,
    refresh_token VARCHAR(1000),
    token_type VARCHAR(20) DEFAULT 'Bearer',
    scopes TEXT COMMENT 'Comma-separated scopes',
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_access_token (access_token),
    INDEX idx_user_company (user_id, company_id),
    INDEX idx_expires_at (expires_at),
    CONSTRAINT fk_tokens_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_tokens_company FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE
);

CREATE TABLE password_reset_tokens (
    id CHAR(36) PRIMARY KEY COMMENT 'UUID',
    user_id CHAR(36) NOT NULL,
    token_hash VARCHAR(64) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    UNIQUE KEY uq_token_hash (token_hash),
    CONSTRAINT fk_reset_tokens_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Billing/subscription data model. No service, handler, or UI wired up yet —
-- schema-only groundwork for recurring billing (e.g. Stripe) later.

CREATE TABLE plans (
    id CHAR(36) PRIMARY KEY COMMENT 'UUID',
    code VARCHAR(50) UNIQUE NOT NULL COMMENT 'free, pro, enterprise',
    name VARCHAR(100) NOT NULL,
    description TEXT,
    price_cents INT NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    billing_interval VARCHAR(20) NOT NULL DEFAULT 'month' COMMENT 'month, year',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_code (code)
);

CREATE TABLE subscriptions (
    id CHAR(36) PRIMARY KEY COMMENT 'UUID',
    company_id CHAR(36) NOT NULL,
    plan_id CHAR(36) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active' COMMENT 'trialing, active, past_due, canceled, incomplete',
    external_provider VARCHAR(50) COMMENT 'e.g. stripe',
    external_subscription_id VARCHAR(255),
    current_period_start TIMESTAMP NULL,
    current_period_end TIMESTAMP NULL,
    cancel_at_period_end BOOLEAN DEFAULT FALSE,
    trial_ends_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_company_id (company_id),
    INDEX idx_status (status),
    INDEX idx_external_subscription_id (external_subscription_id),
    CONSTRAINT fk_subscriptions_company FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE,
    CONSTRAINT fk_subscriptions_plan FOREIGN KEY (plan_id) REFERENCES plans(id)
);

CREATE TABLE invoices (
    id CHAR(36) PRIMARY KEY COMMENT 'UUID',
    company_id CHAR(36) NOT NULL,
    subscription_id CHAR(36),
    external_invoice_id VARCHAR(255),
    amount_cents INT NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    status VARCHAR(20) NOT NULL DEFAULT 'draft' COMMENT 'draft, open, paid, void, uncollectible',
    period_start TIMESTAMP NULL,
    period_end TIMESTAMP NULL,
    paid_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_company_id (company_id),
    INDEX idx_subscription_id (subscription_id),
    INDEX idx_status (status),
    CONSTRAINT fk_invoices_company FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE,
    CONSTRAINT fk_invoices_subscription FOREIGN KEY (subscription_id) REFERENCES subscriptions(id) ON DELETE SET NULL
);

-- Audit trail for sensitive mutations (company create/update/suspend/delete,
-- member invite/role-change/remove). MySQL-only — see service.AuditService;
-- under sqlite these events are just logged to stdout, not persisted here.
CREATE TABLE audit_logs (
    id CHAR(36) PRIMARY KEY COMMENT 'UUID',
    actor_user_id CHAR(36) NOT NULL,
    action VARCHAR(100) NOT NULL COMMENT 'e.g. company.create, member.remove',
    resource_type VARCHAR(50) NOT NULL,
    resource_id CHAR(36) NOT NULL,
    company_id CHAR(36) NULL,
    metadata TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_actor_user_id (actor_user_id),
    INDEX idx_company_id (company_id),
    INDEX idx_resource (resource_type, resource_id),
    INDEX idx_created_at (created_at),
    CONSTRAINT fk_audit_logs_actor FOREIGN KEY (actor_user_id) REFERENCES users(id)
);
