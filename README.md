# redeemer
Twitch redeem tracker
Written in GO

Very early, basically the bare minimum functional code

## A small web site utility to track twitch custom reward redemptions

You will need to ensure environmental variables are set
REDEEM_URL = 'base url for callbacks etc'

REDEEM_DB = 'postgresql://user:pass@host/database'

REDEEM_ID=twitch_dev_app_id

REDEEM_SECRET=twitch_dev_app_secret

REDEEM_SCOPE=channel:read:redemptions+moderator:read:followers

REDEEM_EVENT_SECRET=secret_for_verification

REDEEM_PORT=port_to_host_on

REDEEM_USER=admin_user

REDEEM_PASS=admin_pass



## TODO

- Flesh out error handling
- Add user invite flow
- Add user link flow
- Add table filtering
- Add random picker
- Convert debugging prints to logging
- add csrf protection
- Implement auth middleware instead of manual checks
