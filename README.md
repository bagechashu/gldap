# GLDAP
LDAP server power by golang.

# Notice
Only Mysql is tested, About the backend database.

# Ldap tools

```
ldapsearch -x -D "your_bind_dn" -W -b "search_base" -H ldap://your_ldap_server -s sub "(uid=username)"
ldapsearch -x -D "cn=hackers,dc=gldap,dc=com" -W -b "dc=gldap,dc=com" -H ldap://localhost:389 -s sub "(uid=hackers)"


```

# Add user by "ldapadd"

```
$ cat << EOS > base.ldif

dn: dc=gldap,dc=com
objectClass: top
objectClass: dcObject
objectClass: organization
o: Gldap Inc.
dc: gldap

dn: ou=Users,dc=gldap,dc=com
objectClass: organizationalUnit
ou: Users

dn: ou=Groups,dc=gldap,dc=com
objectClass: organizationalUnit
ou: Group

EOS

$ ldapadd -H ldap://localhost:389 -x -D cn=admin,dc=gldap,dc=com -w secret -f base.ldif
adding new entry "dc=gldap,dc=com"

adding new entry "ou=Users,dc=gldap,dc=com"

adding new entry "ou=Groups,dc=gldap,dc=com"

```

# Thanks List
- https://github.com/glauth/glauth.
