# Deployment Documentation

This directory contains production deployment configurations for self-hosting the Referee Scheduler application.

## Deployment Options

### ✅ Traefik + Cloudflare Tunnel (Recommended)

This setup is configured for use with:
- **Traefik** as reverse proxy
- **Cloudflare Tunnel** for secure ingress
- No need for SSL certificates (Cloudflare handles it)
- Simplified configuration

**Quick Start:** See [QUICK_DEPLOY.md](QUICK_DEPLOY.md)  
**Full Guide:** See [DEPLOYMENT_TRAEFIK.md](DEPLOYMENT_TRAEFIK.md)

### Alternative: Nginx + Let's Encrypt

Files are included for traditional nginx deployment, but **not recommended** if you already have Traefik.

## Files Overview

### Production Docker Compose
- **`docker-compose.prod.yml`** - Production configuration with Traefik labels

### Environment Configuration
- **`.env.production.example`** - Template for production environment
- **`.env.production`** - Your actual config (create from template, not in git)

### Deployment Scripts
- **`scripts/traefik-setup.sh`** - Initial setup wizard
- **`scripts/backup-database.sh`** - Database backup
- **`scripts/restore-database.sh`** - Database restore

### Documentation
- **`QUICK_DEPLOY.md`** - One-page quick reference
- **`DEPLOYMENT_TRAEFIK.md`** - Complete deployment guide
- **`DEPLOYMENT.md`** - Generic deployment info (nginx-based, optional)

### Nginx Configuration (Optional/Reference)
- **`nginx/nginx.conf`** - Main nginx config (not used with Traefik)
- **`nginx/conf.d/ref-sched.conf.template`** - Site config template (not used with Traefik)

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                     Internet                             │
└──────────────────────┬──────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────┐
│              Cloudflare Tunnel                           │
│         (SSL/TLS, DDoS Protection, CDN)                  │
└──────────────────────┬──────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────┐
│                   Traefik                                │
│           (Reverse Proxy, Routing)                       │
│               Network: traefik_default                   │
└───────────┬──────────────────────┬──────────────────────┘
            │                      │
            ▼                      ▼
    ┌───────────────┐      ┌─────────────────┐
    │   Frontend    │      │    Backend      │
    │  SvelteKit    │      │   Go/Gin API    │
    │   Port 3000   │      │   Port 8080     │
    └───────────────┘      └────────┬────────┘
                                    │
                                    ▼
                           ┌─────────────────┐
                           │   PostgreSQL    │
                           │   Port 5432     │
                           │  (Internal Only)│
                           └─────────────────┘
```

## Key Features

### Security
- ✅ Database not exposed to internet
- ✅ All traffic through Cloudflare & Traefik
- ✅ Secure session management
- ✅ Google OAuth for authentication
- ✅ Environment-based secrets

### High Availability
- ✅ Automatic container restart
- ✅ Health checks
- ✅ Cloudflare DDoS protection
- ✅ Automatic SSL renewal (via Cloudflare)

### Data Protection
- ✅ Automated backup scripts
- ✅ Point-in-time restore capability
- ✅ Backup retention policies
- ✅ Database volume persistence

### Observability
- ✅ Container logging
- ✅ Application logs
- ✅ Health check endpoints
- ✅ Traefik dashboard integration

## Quick Deployment

```bash
# 1. Clone repository
git clone <repo-url> /opt/referee-scheduler
cd /opt/referee-scheduler

# 2. Run setup
chmod +x scripts/traefik-setup.sh
./scripts/traefik-setup.sh

# 3. Configure Cloudflare Tunnel
# Add domain to your tunnel config

# 4. Deploy
docker-compose -f docker-compose.prod.yml up -d --build

# 5. Verify
docker-compose -f docker-compose.prod.yml ps
```

See [QUICK_DEPLOY.md](QUICK_DEPLOY.md) for detailed steps.

## Configuration Requirements

Before deploying, you need:

1. **Domain Name**
   - Configured in Cloudflare
   - Cloudflare Tunnel pointing to Traefik

2. **Google OAuth Credentials**
   - Client ID
   - Client Secret
   - Redirect URI configured

3. **Traefik**
   - Running with `traefik_default` network
   - HTTPS entrypoint configured

4. **Server Resources**
   - Minimum: 2GB RAM, 2 CPU cores, 20GB disk
   - Recommended: 4GB RAM, 4 CPU cores, 40GB disk

## Environment Variables

Required in `.env.production`:

```bash
# Domain & URLs
DOMAIN=ref-sched.yourdomain.com
FRONTEND_URL=https://ref-sched.yourdomain.com
VITE_API_URL=https://ref-sched.yourdomain.com/api

# Database
POSTGRES_USER=referee_scheduler
POSTGRES_PASSWORD=<strong-password>
POSTGRES_DB=referee_scheduler

# Security
SESSION_SECRET=<random-secret>

# OAuth
GOOGLE_CLIENT_ID=<your-client-id>
GOOGLE_CLIENT_SECRET=<your-secret>
GOOGLE_REDIRECT_URL=https://ref-sched.yourdomain.com/api/auth/google/callback

# Backups
BACKUP_RETENTION_DAYS=30
```

## Maintenance Tasks

### Daily
- Monitor logs for errors
- Check disk space
- Verify backups completed

### Weekly
- Review backup integrity
- Check container health
- Monitor resource usage

### Monthly
- Update Docker images
- Review security updates
- Test restore procedure
- Rotate logs

### Quarterly
- Full backup verification
- Performance review
- Capacity planning
- Security audit

## Backup & Restore

### Automated Backups

```bash
# Set up daily backups (2 AM)
crontab -e

# Add this line:
0 2 * * * cd /opt/referee-scheduler && ./scripts/backup-database.sh >> /var/log/ref-sched-backup.log 2>&1
```

### Manual Backup

```bash
./scripts/backup-database.sh
```

### Restore

```bash
./scripts/restore-database.sh ./backups/referee_scheduler_YYYYMMDD_HHMMSS.sql.gz
```

## Monitoring

### Container Status
```bash
docker-compose -f docker-compose.prod.yml ps
```

### Live Logs
```bash
docker-compose -f docker-compose.prod.yml logs -f
```

### Resource Usage
```bash
docker stats
```

### Database Health
```bash
docker-compose -f docker-compose.prod.yml exec db pg_isready
```

## Troubleshooting

See detailed troubleshooting in [DEPLOYMENT_TRAEFIK.md](DEPLOYMENT_TRAEFIK.md#troubleshooting)

Common issues:
- Service not accessible → Check Traefik routing
- OAuth fails → Verify redirect URI
- Database connection → Check network configuration
- Container crashes → Review logs

## Cost Estimation

Typical monthly costs for self-hosting:

| Item | Cost |
|------|------|
| VPS (2GB RAM, 2 CPU) | $5-15 |
| Domain Registration | ~$1 |
| Cloudflare (Free tier) | $0 |
| Traefik (self-hosted) | $0 |
| **Total** | **$6-16/month** |

Compare to:
- Azure App Service: $50-100+/month
- AWS ECS: $30-80+/month
- Heroku: $50+/month

## Support & Documentation

- **Quick Start:** [QUICK_DEPLOY.md](QUICK_DEPLOY.md)
- **Full Guide:** [DEPLOYMENT_TRAEFIK.md](DEPLOYMENT_TRAEFIK.md)
- **Main README:** [README.md](README.md)
- **Setup Guide:** [SETUP.md](../guides/SETUP.md)
- **Testing:** [TESTING_GUIDE.md](../guides/TESTING_GUIDE.md)

## Security Considerations

1. **Never commit `.env.production`** to version control
2. **Use strong passwords** (32+ characters)
3. **Rotate secrets** periodically
4. **Keep backups offsite** and encrypted
5. **Update regularly** for security patches
6. **Monitor logs** for suspicious activity
7. **Restrict database access** to Docker network only

## Next Steps After Deployment

1. ✅ Verify application is accessible
2. ✅ Test Google OAuth login
3. ✅ Create first admin user
4. ✅ Import initial match schedule
5. ✅ Configure automated backups
6. ✅ Set up monitoring
7. ✅ Train users
8. ✅ Document your specific setup

---

**Ready to deploy?** Start with [QUICK_DEPLOY.md](QUICK_DEPLOY.md)
