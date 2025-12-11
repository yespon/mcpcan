package postgres

import (
	"database/sql"
	"errors"
	"time"
)

type TenantAccountJoin struct {
	TenantId  string `json:"tenantId,omitempty"`
	AccountId string `json:"accountId,omitempty"`
}

func GetOwnerTenantAccountJoins(db *sql.DB) ([]TenantAccountJoin, error) {
	query := `SELECT tenant_id, account_id FROM tenant_account_joins WHERE role = 'owner'`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tenantAccountJoins []TenantAccountJoin
	for rows.Next() {
		var join TenantAccountJoin
		err := rows.Scan(&join.TenantId, &join.AccountId)
		if err != nil {
			return nil, err
		}
		tenantAccountJoins = append(tenantAccountJoins, join)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tenantAccountJoins, nil
}

type Tenant struct {
	Id               string `json:"id,omitempty"`
	Name             string `json:"name,omitempty"`
	EncryptPublicKey string `json:"encrypt_public_key,omitempty"`
}

func GetAllTenants(db *sql.DB) ([]Tenant, error) {
	query := `SELECT id, name, encrypt_public_key FROM tenants`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tenants []Tenant
	for rows.Next() {
		var tenant Tenant
		err := rows.Scan(&tenant.Id, &tenant.Name, &tenant.EncryptPublicKey)
		if err != nil {
			return nil, err
		}
		tenants = append(tenants, tenant)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tenants, nil
}

type Account struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

func GetAllAccounts(db *sql.DB) ([]Account, error) {
	query := `SELECT id, name FROM accounts`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []Account
	for rows.Next() {
		var account Account
		err := rows.Scan(&account.Id, &account.Name)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return accounts, nil
}

type ToolMcpProvider struct {
	ID                   string    `json:"id,omitempty"`
	Name                 string    `json:"name,omitempty"`
	ServerIdentifier     string    `json:"serverIdentifier,omitempty"`
	ServerURL            string    `json:"serverUrl,omitempty"`
	ServerURLHash        string    `json:"serverUrlHash,omitempty"`
	Icon                 string    `json:"icon,omitempty"`
	TenantId             string    `json:"tenantId,omitempty"`
	UserId               string    `json:"userId,omitempty"`
	EncryptedCredentials string    `json:"encryptedCredentials,omitempty"`
	Authed               bool      `json:"authed,omitempty"`
	Tools                string    `json:"tools,omitempty"`
	Timeout              int       `json:"timeout,omitempty"`
	SseReadTimeout       int       `json:"sse_read_timeout,omitempty"`
	EncryptedHeaders     string    `json:"encrypted_headers,omitempty"`
	CreatedAt            time.Time `json:"createdAt,omitempty"`
	UpdatedAt            time.Time `json:"updatedAt,omitempty"`
}

func GetToolMcpProvider(db *sql.DB, serverURL string, tenantId string) (*ToolMcpProvider, error) {
	query := `SELECT id, name, server_identifier, server_url, server_url_hash, icon, tenant_id, user_id, 
		encrypted_credentials, authed, tools, created_at, updated_at 
		FROM tool_mcp_providers 
		WHERE tenant_id = $1 AND server_url_hash = $2`

	toolMcpProvider := &ToolMcpProvider{}
	err := db.QueryRow(query, tenantId, serverURL).Scan(
		&toolMcpProvider.ID,
		&toolMcpProvider.Name,
		&toolMcpProvider.ServerIdentifier,
		&toolMcpProvider.ServerURL,
		&toolMcpProvider.ServerURLHash,
		&toolMcpProvider.Icon,
		&toolMcpProvider.TenantId,
		&toolMcpProvider.UserId,
		&toolMcpProvider.EncryptedCredentials,
		&toolMcpProvider.Authed,
		&toolMcpProvider.Tools,
		&toolMcpProvider.CreatedAt,
		&toolMcpProvider.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toolMcpProvider, nil
}

func CreateToolMcpProvider(db *sql.DB, toolMcpProvider *ToolMcpProvider) error {
	query := `INSERT INTO tool_mcp_providers 
		(name, server_identifier, server_url, server_url_hash, icon, tenant_id, user_id, 
		encrypted_credentials, authed, tools, created_at, updated_at, timeout, sse_read_timeout, encrypted_headers) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`

	_, err := db.Exec(query,
		toolMcpProvider.Name,
		toolMcpProvider.ServerIdentifier,
		toolMcpProvider.ServerURL,
		toolMcpProvider.ServerURLHash,
		toolMcpProvider.Icon,
		toolMcpProvider.TenantId,
		toolMcpProvider.UserId,
		toolMcpProvider.EncryptedCredentials,
		toolMcpProvider.Authed,
		toolMcpProvider.Tools,
		toolMcpProvider.CreatedAt,
		toolMcpProvider.UpdatedAt,
		toolMcpProvider.Timeout,
		toolMcpProvider.SseReadTimeout,
		toolMcpProvider.EncryptedHeaders,
	)
	return err
}

func UpdateToolMcpProvider(db *sql.DB, toolMcpProvider *ToolMcpProvider) error {
	query := `UPDATE tool_mcp_providers SET 
		name = $1, tools = $2, updated_at = $3 
		WHERE id = $4`

	_, err := db.Exec(query,
		toolMcpProvider.Name,
		toolMcpProvider.Tools,
		toolMcpProvider.UpdatedAt,
		toolMcpProvider.ID,
	)
	return err
}
