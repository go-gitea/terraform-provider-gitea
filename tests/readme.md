# How to run tests

```bash
docker run -p 3000:3000 gitea/gitea:latest-rootless
```

1. Setup gitea under http://localhost:3000
2. Username: gitea_admin
3. Password: gitea_admin
4. Email: admin@gitea.local

```bash
terraform test
```
