// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package database

// SQL schema for WulfVault database

const SchemaVersion = 1

const CreateTablesSQL = `
-- Users table (extended from Gokapi)
CREATE TABLE IF NOT EXISTS Users (
	Id INTEGER PRIMARY KEY AUTOINCREMENT,
	Name TEXT NOT NULL UNIQUE,
	Email TEXT NOT NULL UNIQUE,
	Password TEXT,
	Permissions INTEGER NOT NULL DEFAULT 0,
	Userlevel INTEGER NOT NULL DEFAULT 2,
	LastOnline INTEGER NOT NULL DEFAULT 0,
	ResetPassword INTEGER NOT NULL DEFAULT 0,
	StorageQuotaMB INTEGER NOT NULL DEFAULT 1000,
	StorageUsedMB INTEGER NOT NULL DEFAULT 0,
	CreatedAt INTEGER NOT NULL,
	IsActive INTEGER NOT NULL DEFAULT 1
);

-- Files table (from Gokapi, with UserId for tracking)
CREATE TABLE IF NOT EXISTS Files (
	Id TEXT PRIMARY KEY,
	Name TEXT NOT NULL,
	Size TEXT NOT NULL,
	SHA1 TEXT NOT NULL,
	PasswordHash TEXT,
	FilePasswordPlain TEXT,
	HotlinkId TEXT,
	ContentType TEXT,
	AwsBucket TEXT,
	ExpireAtString TEXT,
	ExpireAt INTEGER,
	PendingDeletion INTEGER DEFAULT 0,
	SizeBytes INTEGER,
	UploadDate INTEGER,
	DownloadsRemaining INTEGER,
	DownloadCount INTEGER DEFAULT 0,
	UserId INTEGER NOT NULL,
	UnlimitedDownloads INTEGER DEFAULT 0,
	UnlimitedTime INTEGER DEFAULT 0,
	RequireAuth INTEGER DEFAULT 0,
	DeletedAt INTEGER DEFAULT 0,
	DeletedBy INTEGER DEFAULT 0,
	FOREIGN KEY (UserId) REFERENCES Users(Id)
);

-- Download Accounts table (for people downloading files with authentication)
CREATE TABLE IF NOT EXISTS DownloadAccounts (
	Id INTEGER PRIMARY KEY AUTOINCREMENT,
	Name TEXT NOT NULL,
	Email TEXT NOT NULL UNIQUE,
	Password TEXT NOT NULL,
	CreatedAt INTEGER NOT NULL,
	LastUsed INTEGER DEFAULT 0,
	DownloadCount INTEGER DEFAULT 0,
	IsActive INTEGER DEFAULT 1
);

-- File Requests table (for requesting file uploads)
CREATE TABLE IF NOT EXISTS FileRequests (
	Id INTEGER PRIMARY KEY AUTOINCREMENT,
	UserId INTEGER NOT NULL,
	RequestToken TEXT NOT NULL UNIQUE,
	Title TEXT NOT NULL,
	Message TEXT,
	CreatedAt INTEGER NOT NULL,
	ExpiresAt INTEGER,
	IsActive INTEGER DEFAULT 1,
	MaxFileSize INTEGER DEFAULT 0,
	AllowedFileTypes TEXT,
	FOREIGN KEY (UserId) REFERENCES Users(Id)
);

-- Download Logs table (tracks all downloads)
CREATE TABLE IF NOT EXISTS DownloadLogs (
	Id INTEGER PRIMARY KEY AUTOINCREMENT,
	FileId TEXT NOT NULL,
	DownloadAccountId INTEGER,
	Email TEXT,
	IpAddress TEXT,
	UserAgent TEXT,
	DownloadedAt INTEGER NOT NULL,
	FileSize INTEGER,
	FileName TEXT,
	IsAuthenticated INTEGER DEFAULT 0,
	FOREIGN KEY (FileId) REFERENCES Files(Id),
	FOREIGN KEY (DownloadAccountId) REFERENCES DownloadAccounts(Id)
);

-- Email Logs table (tracks when files are shared via email)
CREATE TABLE IF NOT EXISTS EmailLogs (
	Id INTEGER PRIMARY KEY AUTOINCREMENT,
	FileId TEXT NOT NULL,
	SenderUserId INTEGER NOT NULL,
	RecipientEmail TEXT NOT NULL,
	Message TEXT,
	SentAt INTEGER NOT NULL,
	FileName TEXT,
	FileSize INTEGER,
	FOREIGN KEY (FileId) REFERENCES Files(Id),
	FOREIGN KEY (SenderUserId) REFERENCES Users(Id)
);

-- Sessions table (from Gokapi)
CREATE TABLE IF NOT EXISTS Sessions (
	Id TEXT PRIMARY KEY,
	UserId INTEGER NOT NULL,
	ValidUntil INTEGER NOT NULL,
	FOREIGN KEY (UserId) REFERENCES Users(Id)
);

-- API Keys table (from Gokapi)
CREATE TABLE IF NOT EXISTS ApiKeys (
	Id TEXT PRIMARY KEY,
	FriendlyName TEXT NOT NULL,
	LastUsed INTEGER,
	Permissions INTEGER NOT NULL,
	UserId INTEGER NOT NULL,
	FOREIGN KEY (UserId) REFERENCES Users(Id)
);

-- Configuration table (stores branding and settings)
CREATE TABLE IF NOT EXISTS Configuration (
	Key TEXT PRIMARY KEY,
	Value TEXT NOT NULL
);

-- Email Provider Configuration table (stores email settings)
CREATE TABLE IF NOT EXISTS EmailProviderConfig (
	Id INTEGER PRIMARY KEY AUTOINCREMENT,
	Provider TEXT NOT NULL UNIQUE,
	IsActive INTEGER DEFAULT 0,
	ApiKeyEncrypted TEXT,
	SMTPHost TEXT,
	SMTPPort INTEGER,
	SMTPUsername TEXT,
	SMTPPasswordEncrypted TEXT,
	SMTPUseTLS INTEGER DEFAULT 1,
	FromEmail TEXT NOT NULL,
	FromName TEXT,
	CreatedAt INTEGER NOT NULL,
	UpdatedAt INTEGER NOT NULL
);

-- Password Reset Tokens table
CREATE TABLE IF NOT EXISTS PasswordResetTokens (
	Id INTEGER PRIMARY KEY AUTOINCREMENT,
	Token TEXT NOT NULL UNIQUE,
	Email TEXT NOT NULL,
	AccountType TEXT NOT NULL,
	ExpiresAt INTEGER NOT NULL,
	Used INTEGER DEFAULT 0,
	CreatedAt INTEGER NOT NULL
);

-- Teams table (for team collaboration)
CREATE TABLE IF NOT EXISTS Teams (
	Id INTEGER PRIMARY KEY AUTOINCREMENT,
	Name TEXT NOT NULL,
	Description TEXT,
	CreatedBy INTEGER NOT NULL,
	CreatedAt INTEGER NOT NULL,
	StorageQuotaMB INTEGER NOT NULL DEFAULT 10240,
	StorageUsedMB INTEGER NOT NULL DEFAULT 0,
	IsActive INTEGER DEFAULT 1,
	FOREIGN KEY (CreatedBy) REFERENCES Users(Id)
);

-- Team Members table (junction table for users and teams)
CREATE TABLE IF NOT EXISTS TeamMembers (
	Id INTEGER PRIMARY KEY AUTOINCREMENT,
	TeamId INTEGER NOT NULL,
	UserId INTEGER NOT NULL,
	Role INTEGER DEFAULT 2,
	JoinedAt INTEGER NOT NULL,
	AddedBy INTEGER,
	FOREIGN KEY (TeamId) REFERENCES Teams(Id) ON DELETE CASCADE,
	FOREIGN KEY (UserId) REFERENCES Users(Id) ON DELETE CASCADE,
	FOREIGN KEY (AddedBy) REFERENCES Users(Id),
	UNIQUE(TeamId, UserId)
);

-- Team Files table (tracks which files are shared to teams)
CREATE TABLE IF NOT EXISTS TeamFiles (
	Id INTEGER PRIMARY KEY AUTOINCREMENT,
	FileId TEXT NOT NULL,
	TeamId INTEGER NOT NULL,
	SharedBy INTEGER NOT NULL,
	SharedAt INTEGER NOT NULL,
	FOREIGN KEY (FileId) REFERENCES Files(Id) ON DELETE CASCADE,
	FOREIGN KEY (TeamId) REFERENCES Teams(Id) ON DELETE CASCADE,
	FOREIGN KEY (SharedBy) REFERENCES Users(Id),
	UNIQUE(FileId, TeamId)
);

-- Indices for performance
CREATE INDEX IF NOT EXISTS idx_files_userid ON Files(UserId);
CREATE INDEX IF NOT EXISTS idx_files_sha1 ON Files(SHA1);
CREATE INDEX IF NOT EXISTS idx_downloadlogs_fileid ON DownloadLogs(FileId);
CREATE INDEX IF NOT EXISTS idx_downloadlogs_accountid ON DownloadLogs(DownloadAccountId);
CREATE INDEX IF NOT EXISTS idx_downloadlogs_downloadedat ON DownloadLogs(DownloadedAt);
CREATE INDEX IF NOT EXISTS idx_emaillogs_fileid ON EmailLogs(FileId);
CREATE INDEX IF NOT EXISTS idx_emaillogs_sentat ON EmailLogs(SentAt);
CREATE INDEX IF NOT EXISTS idx_sessions_userid ON Sessions(UserId);
CREATE INDEX IF NOT EXISTS idx_apikeys_userid ON ApiKeys(UserId);
CREATE INDEX IF NOT EXISTS idx_filerequests_userid ON FileRequests(UserId);
CREATE INDEX IF NOT EXISTS idx_filerequests_token ON FileRequests(RequestToken);
CREATE INDEX IF NOT EXISTS idx_passwordresets_token ON PasswordResetTokens(Token);
CREATE INDEX IF NOT EXISTS idx_passwordresets_email ON PasswordResetTokens(Email);
CREATE INDEX IF NOT EXISTS idx_team_members_team ON TeamMembers(TeamId);
CREATE INDEX IF NOT EXISTS idx_team_members_user ON TeamMembers(UserId);
CREATE INDEX IF NOT EXISTS idx_team_files_team ON TeamFiles(TeamId);
CREATE INDEX IF NOT EXISTS idx_team_files_file ON TeamFiles(FileId);
`
