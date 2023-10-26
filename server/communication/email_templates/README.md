# Email Templates

monetr generates email templates in both an html format and a plain text format using react-email. This is done via
CMake at build time. If you are missing the email templates for whatever reason try running:

```shell
make email
```

This will build the generate email template target.

**NOTE**: Generated emails are not version controlled and should not be committed.
