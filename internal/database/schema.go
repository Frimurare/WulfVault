package database

// SQL schema for Sharecare database

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

-- Indices for performance
CREATE INDEX IF NOT EXISTS idx_files_userid ON Files(UserId);
CREATE INDEX IF NOT EXISTS idx_files_sha1 ON Files(SHA1);
CREATE INDEX IF NOT EXISTS idx_downloadlogs_fileid ON DownloadLogs(FileId);
CREATE INDEX IF NOT EXISTS idx_downloadlogs_accountid ON DownloadLogs(DownloadAccountId);
CREATE INDEX IF NOT EXISTS idx_downloadlogs_downloadedat ON DownloadLogs(DownloadedAt);
CREATE INDEX IF NOT EXISTS idx_sessions_userid ON Sessions(UserId);
CREATE INDEX IF NOT EXISTS idx_apikeys_userid ON ApiKeys(UserId);
CREATE INDEX IF NOT EXISTS idx_filerequests_userid ON FileRequests(UserId);
CREATE INDEX IF NOT EXISTS idx_filerequests_token ON FileRequests(RequestToken);
`
