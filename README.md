# ![Logo](frontend/public/favicon.svg) PlainPage

[![Release](https://img.shields.io/github/v/release/tfabritius/plainpage)](https://github.com/tfabritius/plainpage/releases/latest)

PlainPage is a self-hosted wiki app that embodies simplicity, ease of use, and privacy. It offers a plain and simplistic user experience while providing essential wiki-like functionality. It adheres to the 😘 KISS principle, focusing on delivering a straightforward and efficient user experience without unnecessary complexities.

Powered by Go, Vue/Nuxt. Made with 💖

## Table of Contents

- [Key Features](#key-features)
- [Installation and Setup](#installation-and-setup)
  - [Quick Start](#quick-start)
  - [Production Setup](#production-setup)
  - [Configuration](#configuration)
- [Usage](#usage)
  - [Pages and Folders](#pages-and-folders)
  - [Access Rights](#access-rights)
  - [Keyboard Shortcuts](#keyboard-shortcuts)
- [Data Storage](#data-storage)
  - [Directory Structure](#directory-structure)
  - [Version History (Attic)](#version-history-attic)
  - [Trash](#trash)
- [Security](#security)
- [Contributing](#contributing)
  - [Development Setup](#development-setup)
  - [Building from Source](#building-from-source)
  - [Running Tests](#running-tests)

## Key Features

- **Markdown Syntax** ✨ Create and edit your content using Markdown syntax, a lightweight and intuitive markup language that makes writing and formatting a breeze.

- **Plain File Storage** 🤩 PlainPage utilizes a plain file storage system, your content is stored in a simple and accessible format.

- **Snappy User Interface** 🚀 PlainPage offers a snappy and responsive user interface, allowing you to focus on your content and enhancing your productivity.

- **Privacy-First** 🛑 PlainPage does not track your usage, collect statistics, or employ any analytics tools. Your data remains private and under your control.

- **Access Rights** 🔐 PlainPage provides robust access control with powerful yet user-friendly ACLs. You can easily manage and assign permissions to your content, ensuring that the right people have access to the right information.

### More Features

- Localization: English, German, Spanish
- Supports light/dark mode
- Embedded full text search

### Known Limitations

- Not optimized for access by machines, e.g., search engines
- Concurrent/conflicting edits are not prevented

## Installation and Setup

### Quick Start

Trying out PlainPage is as simple as it can be:

**Option 1: Download Binary**

1. Download the [latest release](https://github.com/tfabritius/plainpage/releases/latest),
2. Run the PlainPage executable, and
3. Browse to <http://localhost:8080/>

**Option 2: Docker**

```bash
docker run --rm -p 8080:8080 ghcr.io/tfabritius/plainpage
```

⚠️ **Warning:** This quick start setup is not production ready. Data will not be persisted.

### Production Setup

#### Using Docker Compose (Recommended)

Create a `docker-compose.yml` file:

```yaml
services:
  plainpage:
    image: ghcr.io/tfabritius/plainpage:latest
    container_name: plainpage
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - ./data:/data
```

Start the service:

```bash
docker compose up -d
```

#### Using Docker Run

Make sure to persist your data by mounting a volume or local folder to the container's `/data` directory:

```bash
docker run -d \
  --name plainpage \
  --restart unless-stopped \
  -p 8080:8080 \
  -v /path/on/host:/data \
  ghcr.io/tfabritius/plainpage
```

#### Reverse Proxy

When running in production you typically want to run a reverse proxy in front of PlainPage that provides encryption via TLS/SSL. Popular choices are:

- [Caddy](https://caddyserver.com/)
- [Traefik](https://traefik.io/)
- [Nginx](https://nginx.org/)
- [Apache](https://httpd.apache.org/)

💡 **Tip:** As usual, backup your data regularly!

### Configuration

PlainPage executable has very few configuration settings with sensible default values that can be configured by environment variables (or a `.env` file):

```ini
# Directory to store data, defaults to `data` directory next to executable
DATA_DIR=./data

# Port to listen on, defaults to 8080
PORT=8080
```

All other settings can be done via the UI or by editing the `config.yml` file in the data directory:

```yaml
# Title of the app - customize your installation
appTitle: "My app"

# Access rules independent from individual pages/folders
acl: []

# Secret for signing JWT tokens (auto-generated, keep it safe!)
jwtSecret: "..."

# Enables anonymous registration with admin rights (auto-disabled after first registration)
setupMode: false

# Retention policies for automatic cleanup (all values default to 0 = disabled)
retention:
  trash:
    maxAgeDays: 30    # Delete trash items older than 30 days
  attic:
    maxAgeDays: 90    # Delete versions older than 90 days
    maxVersions: 50   # Keep at most 50 versions per page
```

⚠️ **Security Note:** The `jwtSecret` is used to sign and verify JWT tokens. It is generated automatically. Keep it safe! For security reasons it's neither exposed nor can be changed via UI.

#### Retention Policies

PlainPage can automatically clean up old trash items and version history to manage disk space. Retention policies are configured via UI and stored in `config.yml`.

**Notes:**
- All retention settings default to `0` (disabled) for safety
- Cleanup runs automatically every 24 hours
- For attic cleanup, versions are deleted if *either* the age limit *or* the version count limit is exceeded

## Usage

### Pages and Folders

PlainPage stores your information on pages. Pages are organized into folders that can be nested.

Pages are written in Markdown syntax with support for:

- Headings, paragraphs, and text formatting
- Lists (ordered and unordered)
- Links and images
- Code blocks with syntax highlighting
- Tables

### Access Rights

Access rights can be modified by administrators.

#### Access Rights for Pages and Folders

PlainPage's permission system is powerful yet easy to use. If you don't adjust it, all registered users will be able to access and edit all content.

Permissions can be defined for both pages and folders. By default, permissions for pages and folders are inherited from the parent folder.

- A page `/folder/page` will inherit its access rights from the folder `/folder`, unless explicit permissions are defined for this page.
- The folder `/folder` will itself inherit its permissions from the parent folder `/`, unless permissions are defined for `/folder`.
- The home folder `/` cannot inherit permissions.

Permissions can be granted to:
- Individual users
- All registered users
- Anonymous users (not logged in)

Permissions are controlled with fine-grained operations:

| Permission | Description                                                        |
| ---------- | ------------------------------------------------------------------ |
| *Read*     | Allows users to access pages and folders                           |
| *Write*    | Allows users to change pages and folders (e.g., create new ones)   |
| *Delete*   | Allows users to delete pages and folders completely                |

This allows you to restrict certain content to specific users and/or expose certain content publicly.

#### Access Rights Beyond Pages and Folders

Besides pages and folders, PlainPage allows you to grant additional permissions:

- **Register** – Allows creating new users in PlainPage. This way you can control whether new users can create their own account and whether existing users can create additional user accounts. By default, only administrators can register new users.

- **Admin** – Grants special rights, e.g., to change permissions. Users with this privilege are automatically granted all other possible permissions on all content.

### Keyboard Shortcuts

| Shortcut           | Action                                          |
| ------------------ | ----------------------------------------------- |
| `Ctrl+K`, `/`      | Search                                          |
| `E`                | Edit page                                       |
| `Ctrl+S`           | Save page (when editing)                        |
| `Esc`              | Cancel edit / close full screen mode            |
| `Alt+↑`            | Navigate to parent folder                       |
| `Ctrl+Backspace`   | Delete current page/folder                      |

## Data Storage

PlainPage uses a plain file storage system. All data is stored in a configurable data directory (default: `data/`).

### Directory Structure

```
data/
├── config.yml          # Application configuration
├── users.yml           # User accounts
├── pages/              # Current pages and folders
│   ├── _index.md       # Root folder metadata
│   ├── mypage.md       # Page at /mypage
│   └── docs/           # Folder at /docs
│       ├── _index.md   # Folder metadata
│       └── page.md     # Page at /docs/page
├── attic/              # Version history
│   ├── mypage.1707740000.md      # Version of /mypage
│   └── docs/
│       └── page.1707745000.md    # Version of /docs/page
└── trash/              # Deleted pages
    └── docs/
        └── guide/
            └── _1707750000/      # Deletion timestamp (prefixed with _)
                ├── guide.md              # Deleted page
                ├── guide.1707740000.md   # Attic entries at deletion
                └── guide.1707745000.md
```

### Pages and Folders

- **Pages** are stored as Markdown files with YAML frontmatter (`.md`)
- **Folders** are directories containing an `_index.md` file with folder metadata
- The frontmatter contains metadata like title, ACL, and modification info

### Version History (Attic)

Every time a page is saved, a copy is stored in the `attic/` directory with the same path structure. The filename includes a Unix timestamp: `{pagename}.{timestamp}.md`
The attic also contains the current version.

💡 **Tip:** Configure [retention policies](#retention-policies) to automatically clean up old versions and manage disk space.

### Trash

When pages are deleted, they are moved to the `trash/` directory instead of being permanently deleted. This allows for recovery if needed.

**Trash structure:** `trash/{original-url-path}/{deletion-timestamp}/`

Each deletion creates a timestamped folder containing:
- The deleted page file
- All attic entries that existed at the time of deletion

If a page was deleted multiple times, each deletion has its own timestamp folder, making it easy to see the history and choose which version to restore.

💡 **Tip:** Configure [retention policies](#retention-policies) to automatically clean up old trash items and free up disk space.

## Security

PlainPage takes security seriously:

- **Password Storage** – User passwords are hashed using [Argon2](https://en.wikipedia.org/wiki/Argon2), the winner of the Password Hashing Competition.
- **Authentication** – JWT (JSON Web Tokens) are used for session management.
- **No Tracking** – No analytics, telemetry, or external requests are made.

### Security Best Practices

1. Always run PlainPage behind a reverse proxy with TLS/SSL in production
2. Keep the `jwtSecret` in `config.yml` secure and backed up
3. Use strong passwords for user accounts
4. Regularly backup your data directory
5. Keep PlainPage updated to the latest version

## Contributing

🤩 You're welcome to contribute to PlainPage!

### Development Setup

During development, there are two processes running:

**1. Nuxt Frontend**

```bash
cd frontend
pnpm install
pnpm dev
```

**2. Go Backend**

```bash
cd backend
go run .
```

Browse to [http://localhost:3000](http://localhost:3000). Nuxt will proxy requests to the backend.

### Building from Source

In production, the static frontend files are served together with the backend:

**1. Generate static frontend files**

```bash
cd frontend
pnpm install
pnpm generate
```

**2. Build Go backend (with embedded frontend)**

```bash
cd backend
go generate ./...
go build .
```

The resulting binary can be found in the `backend` directory. Browse to [http://localhost:8080](http://localhost:8080).

### Running Tests

**Backend Tests**

```bash
cd backend
go test ./...
```

**Frontend Linting and Typechecking**

```bash
cd frontend
pnpm lint
pnpm typecheck
```

### Contributing Guidelines

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests to ensure everything works
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

---

If you find PlainPage useful, please consider giving it a ⭐ on GitHub!
