# Function which triggers emails on Oracle Autonomous Transaction Processing Database events

This function will send email notification (to a configurable email addresses) when a new instance of [Oracle ATP Database](https://docs.cloud.oracle.com/iaas/Content/Database/Concepts/atpoverview.htm) is created (with a specific tag) and also after the instance provisioning completes. Notifications are powered by [Oracle Cloud Infrastructure Email Delivery](https://docs.cloud.oracle.com/iaas/Content/Email/Concepts/overview.htm). It's a Go function which uses the [SMTP package in Go](https://golang.org/pkg/net/smtp/) to send emails

For example, as soon as the instance creation starts, you'll get an email with a subject and body which similar to below

**Subject**

ATP Database instance DB FOOBAR in status PROVISIONING

**Body**

ATP Database instance DB FOOBAR in status PROVISIONING
Instance OCID: ocid1.autonomousdatabase.oc1.phx.abyhqljt4i2hwwtg5l7nkpeu5bm5hx42dtegb7wxvykswl4q4lsmxwuudevq

## Pre-requisites

### Configure OCI Email Delivery

- [Generate SMTP Credentials for a User](https://docs.cloud.oracle.com/iaas/Content/Email/Tasks/generatesmtpcredentials.htm) - you'll have to configure these user credentials in the app (`OCI_EMAIL_DELIVERY_USER_OCID` and `OCI_EMAIL_DELIVERY_USER_PASSWORD` variables)
- [Add approved sender](https://docs.cloud.oracle.com/iaas/Content/Email/Tasks/managingapprovedsenders.htm) - use `OCI_EMAIL_DELIVERY_APPROVED_SENDER` parameter to configure this in the app
- [Note down value for the SMTP server](https://docs.cloud.oracle.com/iaas/Content/Email/Tasks/configuresmtpconnection.htm) - it'll be used in the `OCI_EMAIL_DELIVERY_SMTP_SERVER` configuration attribute

Clone this repo - `git clone https://github.com/abhirockzz/fn-atp-event-notification-app`

### Switch to correct context

- `fn use context <your context name>`
- Check using `fn ls apps`

### Create app

`fn create app --annotation oracle.com/oci/subnetIds=<SUBNETS> --config OCI_EMAIL_DELIVERY_USER_OCID=<OCI_EMAIL_DELIVERY_USER_OCID> --config OCI_EMAIL_DELIVERY_USER_PASSWORD=<OCI_EMAIL_DELIVERY_USER_PASSWORD> --config OCI_EMAIL_DELIVERY_SMTP_SERVER=<OCI_EMAIL_DELIVERY_SMTP_SERVER> --config OCI_EMAIL_DELIVERY_APPROVED_SENDER=<OCI_EMAIL_DELIVERY_APPROVED_SENDER> --config EMAIL_NOTIFICAITON_RECEPIENT_ADDRESS=<EMAIL_NOTIFICAITON_RECEPIENT_ADDRESS> fn-atp-event-notification-app`

> Please provide a valid email address for `EMAIL_NOTIFICAITON_RECEPIENT_ADDRESS`. It is also possible to provide multiple email addresses separated by a comma (`,`) e.g. `foo@gmail.com,bar@gmail.com`

e.g.

`fn create app --annotation oracle.com/oci/subnetIds='["ocid1.subnet.oc1.phx.aaaaaaaabrg4uf2uzc3ni4jkz5vhqwprofmlmo7mpumnuddd7iandssruohq"]' --config OCI_EMAIL_DELIVERY_USER_OCID=ocid1.user.oc1..aaaaaaaa4seqx6jeyma46ldy4cbuv35q4l26scz5p4rkz3rauuoioo26qwmq@ocid1.tenancy.oc1..aaaaaaaaydrjm77otncda2xn7qtv7l3hqnd3zxn2u6siwdhniibwfv4wwhta.3n.com --config OCI_EMAIL_DELIVERY_USER_PASSWORD='foobar' --config OCI_EMAIL_DELIVERY_SMTP_SERVER=smtp.us-phoenix-1.oraclecloud.com --config OCI_EMAIL_DELIVERY_APPROVED_SENDER=test@test.com --config EMAIL_NOTIFICAITON_RECEPIENT_ADDRESS=abhirockzz@gmail.com fn-atp-event-notification-app`

> Ensure that the value for `OCI_EMAIL_DELIVERY_USER_PASSWORD` is surrounded by `''` e.g. `'foO$>>Ba-rG2E{DiN3)d'`

**Check**

`fn inspect app fn-atp-event-notification-app`

## Deploy the app

`cd fn-atp-event-notification-app` and `fn -v deploy --app fn-atp-event-notification-app`

## Test

### Standalone

To test without end-to-end Events integration, just simulate the instance creation and completion events by using the sample payloads included in the source to invoke the function *manually*

- To test instance **creation** event - `cat atp-create-start.json | fn invoke fn-atp-event-notification-app notifyonevent`

- To test instance creation **completion** event - `cat atp-create-end.json | fn invoke fn-atp-event-notification-app notifyonevent`

You should recieve the emails

### Events integration

**Create Events rule**

Before creating the rule, find the function OCID (use the command below) and replace it in `functionId` attribute in the `actions.json` file 

`fn inspect fn fn-atp-event-notification-app notifyonevent | jq '.id' | sed -e 's/^"//' -e 's/"$//'`

Go ahead and create the rule... 

`oci --profile <oci-config-profile-name> cloud-events rule create --display-name <display-name> --is-enabled true --condition '{"eventType":["com.oraclecloud.dbaas.autonomous.database.instance.create.begin","com.oraclecloud.dbaas.autonomous.database.instance.create.end"],"data":{"freeformTags":{<custom-tag-key-value-pair>}}}' --compartment-id <compartment-ocid> --actions file://<filename>.json`

Replace `<custom-tag-key-value-pair>` with the tag you want to use while creating the ATP instance (details below)

e.g.

`oci --profile my-oci-profile cloud-events rule create --display-name invoke-function-on-atp-events --is-enabled true --condition '{"eventType":["com.oraclecloud.dbaas.autonomous.database.instance.create.begin","com.oraclecloud.dbaas.autonomous.database.instance.create.end"],"data":{"freeformTags":{"created_by":"foobar"}}}' --compartment-id ocid1.compartment.oc1..aaaaaaaaokbzj2jn3hf5kwdwqoxl2dq7u54p3tsmxrjd7s3uu7x23tkegiua --actions file://actions.json`

> Note `"created_by":"foobar"` is used as the value in `"freeformTags"`. You can change this if you want to. Jsut ensure that you use the correct tag while creating the Oracle ATP DB instance

**Provision Oracle ATP Database instance**

After setting up Events integration, all you need is to create an [Oracle ATP Database instance](https://docs.cloud.oracle.com/iaas/Content/Database/Tasks/atpcreating.htm)(with the tag you configured while creating the rule)

You should see the email notifications soon....
