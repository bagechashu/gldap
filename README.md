# GLDAP
LDAP server power by golang.

# Notice
Only Mysql is tested, About the backend database.


# Add test data
> https://glauth.github.io/docs/databases.html

```
-- Inserting data into 'users' table
INSERT INTO users (name, uidnumber, primarygroup, givenname, sn, mail, loginshell, homedirectory, disabled, passsha256, passbcrypt, otpsecret, sshkeys, custattr)
VALUES ('hackers', 5001, 5501, 'John', 'Doe', 'john@example.com', '/bin/bash', '/home/hackers', 0, '6478579e37aff45f013e14eeb30b3cc56c72ccdc310123bcdf53e0333e3f416a', '', '', 'ssh-rsa public_key_here', '{"attribute1": "value1", "attribute2": "value2"}');

INSERT INTO users (name, uidnumber, primarygroup, givenname, sn, mail, loginshell, homedirectory, disabled, passsha256, passbcrypt, otpsecret, sshkeys, custattr)
VALUES ('johndoe', 5002, 5502, 'John', 'Doe', 'johndoe@example.com', '/bin/bash', '/home/johndoe', 0, '6478579e37aff45f013e14eeb30b3cc56c72ccdc310123bcdf53e0333e3f416a', '', '', 'ssh-rsa public_key_here', '{"attribute1": "value1", "attribute2": "value2"}');

-- Inserting data into 'ldapgroups' table
INSERT INTO ldapgroups (name, gidnumber) VALUES ('superheros', 5501);
INSERT INTO ldapgroups (name, gidnumber) VALUES ('svcaccts', 5502);
INSERT INTO ldapgroups (name, gidnumber) VALUES ('civilians', 5503);

-- Inserting data into 'includegroups' table
INSERT INTO includegroups (parentgroupid, includegroupid) VALUES (5503, 5501);
INSERT INTO includegroups (parentgroupid, includegroupid) VALUES (5504, 5502);
INSERT INTO includegroups (parentgroupid, includegroupid) VALUES (5504, 5501);

-- Inserting data into 'capabilities' table
INSERT INTO capabilities (userid, action, object) VALUES (5001, 'search', 'ou=superheros,dc=glauth,dc=com');
INSERT INTO capabilities (userid, action, object) VALUES (5003, 'search', '*');

```

Use ldapsearch to test the connection:
```
$ ldapsearch -x -H ldap://localhost:389 -D "cn=hackers,ou=superheros,ou=users,dc=gldap,dc=com" -w dogood -b "ou=superheros,dc=gldap,dc=com" -s sub "(cn=hackers)"
$ ldapsearch -x -H ldap://localhost:389 -D "cn=hackers,ou=superheros,ou=users,dc=gldap,dc=com" -w dogood -b "ou=superheros,dc=gldap,dc=com" -s sub "(uid=hackers)"
$ ldapsearch -x -H ldap://localhost:389 -D "cn=hackers,ou=superheros,dc=gldap,dc=com" -w dogood -b "ou=superheros,dc=gldap,dc=com" -s sub "(uid=hackers)"

#####
# result
#####

# extended LDIF
#
# LDAPv3
# base <ou=superheros,dc=gldap,dc=com> with scope subtree
# filter: (uid=hackers)
# requesting: ALL
#

# hackers, superheros, gldap.com
dn: cn=hackers,ou=superheros,dc=gldap,dc=com
cn: hackers
uid: hackers
givenName: John
sn: Doe
ou: superheros
uidNumber: 5001
accountStatus: active
mail: john@example.com
userPrincipalName: john@example.com
objectClass: posixAccount
loginShell: /bin/bash
homeDirectory: /home/hackers
description: hackers via LDAP
gecos: hackers via LDAP
gidNumber: 5501
memberOf: ou=civilians,ou=groups,dc=gldap,dc=com
memberOf: ou=superheros,ou=groups,dc=gldap,dc=com
sshPublicKey: ssh-rsa public_key_here

# search result
search: 2
result: 0 Success

# numResponses: 2
# numEntries: 1

```

# (X Failed) Add user by "ldapadd"

```
$ cat << EOF > base.ldif
dn: uid=test,ou=superheros,dc=gldap,dc=com
objectClass: top
objectClass: person
objectClass: organizationalPerson
objectClass: inetOrgPerson
cn: test
uid: test
givenName: test
sn: tt
mail: test@example.com
loginShell: /bin/bash
homeDirectory: /home/test
userPassword: {SHA}6478579e37aff45f013e14eeb30b3cc56c72ccdc310123bcdf53e0333e3f416a
gidNumber: 5501
memberOf: ou=superheros,ou=groups,dc=gldap,dc=com
memberOf: ou=civilians,ou=groups,dc=gldap,dc=com
EOF
$ ldapadd -H ldap://localhost:389 -x -D cn=hackers,dc=gldap,dc=com -w dogood -f base.ldif

$ ldapadd -x -D "cn=hackers,dc=gldap,dc=com" -w dogood <<EOF
dn: uid=test,ou=superheros,dc=gldap,dc=com
objectClass: top
objectClass: person
objectClass: organizationalPerson
objectClass: inetOrgPerson
cn: test
uid: test
givenName: test
sn: tt
mail: test@example.com
loginShell: /bin/bash
homeDirectory: /home/test
userPassword: {SHA}6478579e37aff45f013e14eeb30b3cc56c72ccdc310123bcdf53e0333e3f416a
gidNumber: 5501
memberOf: ou=superheros,ou=groups,dc=glauth,dc=com
memberOf: ou=civilians,ou=groups,dc=glauth,dc=com
EOF

```

# Thanks List
- https://github.com/glauth/glauth.
