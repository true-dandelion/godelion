package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID             string         `gorm:"primaryKey" json:"id"`
	Username       string         `gorm:"uniqueIndex;not null" json:"username"`
	PasswordHash   string         `gorm:"not null" json:"-"`
	Role           string         `gorm:"not null" json:"role"` // "admin" or "user"
	SessionTimeout int            `json:"session_timeout"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
	Containers     []Container    `gorm:"foreignKey:UserID" json:"containers,omitempty"`
}

type Container struct {
	ID             string         `gorm:"primaryKey" json:"id"`
	DockerID       string         `json:"docker_id"` // Actual Docker Container ID
	Name           string         `json:"name"`
	Image          string         `json:"image"`
	UserID         string         `json:"user_id"`
	Ports          string         `gorm:"type:text" json:"ports"` // JSON serialized port mappings
	ResourceLimits string         `json:"resource_limits"`
	DeploymentLogs string         `gorm:"type:text" json:"deployment_logs"` // Stores logs during the async creation phase
	Status         string         `gorm:"-" json:"status"` // Transient, fetched from Docker
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
	GatewayRules   []GatewayRule  `gorm:"foreignKey:ContainerID" json:"gateway_rules,omitempty"`
}

type GatewayRule struct {
	ID          string         `gorm:"primaryKey" json:"id"`
	Domain      string         `gorm:"uniqueIndex;not null" json:"domain"`
	ListenPorts string         `json:"listen_ports"` // e.g. "80, 443"
	TargetURLs  string         `json:"target_urls"`  // e.g. "127.0.0.1:3000, demo:3000"
	TargetPort  int            `json:"target_port"`  // Legacy
	ContainerID string         `json:"container_id"` // Legacy
	TLSEnabled  bool           `json:"tls_enabled"`
	CertPath    string         `json:"cert_path"`
	KeyPath     string         `json:"key_path"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type AuditLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    string    `json:"user_id"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource"`
	IPAddress string    `json:"ip_address"`
	Details   string    `json:"details"`
	CreatedAt time.Time `json:"created_at"`
}
