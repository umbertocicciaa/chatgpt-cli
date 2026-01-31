# Security Policy

## Overview

The security of ChatGPT CLI is a top priority. This document outlines our security practices, supported versions, and how to report vulnerabilities.

## Supported Versions

We provide security updates for the following versions:

| Version | Supported          | Notes                          |
| ------- | ------------------ | ------------------------------ |
| main    | :white_check_mark: | Latest development version     |
| latest  | :white_check_mark: | Most recent stable release     |

**Note:** As this project is in active development, we recommend always using the latest version to ensure you have the most recent security patches and improvements.

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security issue, please follow these steps:

### How to Report

You can report security vulnerabilities through:

1. **GitHub Issues**: Create an issue describing the vulnerability (recommended for transparency)
2. **Security Advisory**: Use GitHub's [Security Advisory feature](https://github.com/umbertocicciaa/chatgpt-cli/security/advisories/new) for private disclosure
3. **Email**: Send a detailed report to the repository owner at the email address listed in the [GitHub profile](https://github.com/umbertocicciaa)

### What to Include

Please include the following information in your report:

- **Description**: A clear description of the vulnerability
- **Impact**: The potential impact and severity of the issue
- **Reproduction Steps**: Detailed steps to reproduce the vulnerability
- **Version**: The version of ChatGPT CLI affected
- **Proof of Concept**: If applicable, provide PoC code or screenshots
- **Suggested Fix**: If you have ideas for how to fix it (optional)

### What to Expect

- **Acknowledgment**: You will receive an acknowledgment within **48 hours**
- **Updates**: We will provide regular updates on the progress (at least every 5 business days)
- **Resolution Timeline**: We aim to resolve critical vulnerabilities within 7-14 days
- **Credit**: If you wish, we will credit you in the security advisory and release notes

### Vulnerability Acceptance

If the vulnerability is accepted:
- We will work on a fix and keep you informed of progress
- A security advisory will be published after the fix is released
- You will be credited for the discovery (unless you prefer to remain anonymous)

If the vulnerability is declined:
- We will provide a detailed explanation of why it was not accepted
- You may request a second review if you believe the decision was incorrect

## Security Best Practices for Users

### API Key Security

:warning: **CRITICAL**: Your OpenAI API key is sensitive and should be protected

**DO:**
- ✅ Use environment variables to store your API key (`export OPENAI_API_KEY="sk-..."`)
- ✅ Add `.env` files to `.gitignore` if you use them
- ✅ Use file permissions to protect config files (`chmod 600 ~/.chatgpt-cli/config`)
- ✅ Rotate your API key regularly
- ✅ Use separate API keys for different environments (development, production)
- ✅ Monitor your OpenAI API usage at https://platform.openai.com/account/usage

**DO NOT:**
- ❌ Hard-code API keys in scripts or code
- ❌ Commit API keys to version control (`.git`, `.env`)
- ❌ Share API keys in public forums, chat, or screenshots
- ❌ Store API keys in plain text files with broad read permissions
- ❌ Use production API keys for testing or development

### Configuration File Security

The application stores configuration in `~/.chatgpt-cli/`. This directory may contain sensitive information.

After installing via `go install`, ensure proper permissions on your config files:

```bash
# Ensure proper permissions on config directory
chmod 700 ~/.chatgpt-cli/

# Ensure proper permissions on config file
chmod 600 ~/.chatgpt-cli/config

# Ensure proper permissions on log file (contains prompts/responses)
chmod 600 ~/.chatgpt-cli/logs.jsonl
```

### Network Security

- **HTTPS Only**: The application uses HTTPS for all API communications by default
- **Custom API URLs**: If you set a custom `OPENAI_API_URL`, ensure it uses HTTPS
- **Proxies**: If using a proxy, ensure it supports HTTPS and is trustworthy
- **Network Monitoring**: Be aware that network administrators may be able to see your API requests

### Log File Security

The application logs all prompts and responses to `~/.chatgpt-cli/logs.jsonl`:

- These logs may contain sensitive information from your conversations
- Ensure appropriate file permissions (see above)
- Regularly review and clean up old logs
- Consider the sensitivity of prompts before using the tool in shared environments

### Dependency Security

This project uses Go's standard library exclusively, minimizing external dependencies:

- No third-party dependencies in `go.mod`
- Regular updates to supported Go versions
- Security scanning via GitHub's CodeQL

Users should:
- Keep Go updated to the latest stable version
- Build from source or verify checksums of binary releases
- Review the source code before building (it's intentionally small and readable)

### Multi-User Environments

If running in a multi-user environment:

1. **Use per-user configurations**: Each user should have their own `~/.chatgpt-cli/` directory
2. **Avoid shared API keys**: Each user should use their own OpenAI API key
3. **Log file privacy**: Ensure log files are only readable by the user (`chmod 600`)
4. **Temp directory cleanup**: The application doesn't create temp files, but always verify

### Rate Limiting and Abuse Prevention

- **Respect OpenAI's rate limits**: Excessive requests may result in API key suspension
- **Monitor usage**: Check your OpenAI dashboard regularly
- **Set reasonable timeouts**: Use `OPENAI_TIMEOUT` to prevent hanging requests
- **Don't automate excessive queries**: Avoid scripts that send thousands of requests

## Known Security Considerations

### Current Implementation

1. **API Key Storage**: API keys can be stored in:
   - Environment variables (recommended)
   - Config file at `~/.chatgpt-cli/config` (ensure proper permissions)

2. **Config File Permissions**: The application creates config files with `0600` permissions (owner read/write only), which is appropriate for sensitive data.

3. **Log File Permissions**: Log files are created with `0644` permissions. Users should manually restrict these to `0600` if logs contain sensitive information.

4. **No Encryption at Rest**: API keys and logs are stored in plain text on disk. Users should ensure filesystem-level encryption if this is a concern.

5. **Memory Safety**: Go provides memory safety by default, reducing the risk of buffer overflows and similar vulnerabilities.

### Recommendations for Enhanced Security

For users with heightened security requirements:

1. **Use system keyring**: Consider using a system keyring service (e.g., keychain on macOS, gnome-keyring on Linux) to store API keys instead of plain text
2. **Encrypt logs**: If logs contain highly sensitive data, consider encrypting the logs.jsonl file
3. **Use ephemeral environments**: Run the tool in containers or VMs that are destroyed after use
4. **Audit trail**: Implement additional logging or monitoring for compliance requirements
5. **Network isolation**: Run in a network-isolated environment if handling sensitive data

## Security Updates and Patching

- Security updates are released as soon as possible after a vulnerability is confirmed
- Critical vulnerabilities are prioritized and addressed within 7-14 days
- All security updates are announced in:
  - GitHub Security Advisories
  - Release notes
  - Repository README badges

To stay informed:
- Watch this repository for security advisories
- Check the [Security tab](https://github.com/umbertocicciaa/chatgpt-cli/security) regularly
- Enable GitHub notifications for security alerts

## Compliance and Auditing

This project:
- ✅ Uses GitHub's Dependabot for dependency vulnerability scanning
- ✅ Runs CodeQL analysis for code security issues
- ✅ Follows secure coding practices
- ✅ Maintains minimal external dependencies (none currently)
- ✅ Uses HTTPS for all external communications
- ✅ Implements proper error handling to prevent information disclosure

## Additional Resources

- [OpenAI API Security Best Practices](https://platform.openai.com/docs/guides/production-best-practices)
- [OWASP Secure Coding Practices](https://owasp.org/www-project-secure-coding-practices-quick-reference-guide/)
- [Go Security Policy](https://go.dev/security/policy)

## Questions or Concerns?

If you have questions about security practices or concerns that don't constitute a vulnerability, please:
- Open a GitHub issue (for non-sensitive topics)
- Start a discussion in the GitHub Discussions tab
- Contact the maintainer through GitHub

---

**Last Updated**: January 2026  
**Version**: 1.0
