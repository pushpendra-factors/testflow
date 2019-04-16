# FactorsAI - Javascript SDK

* Intialize your SDK with YOUR_TOKEN once and use other methods as listed below.
```
factorsai.init('<YOUR_TOKEN>');
```

* Tracking a custom event with properties.
```
factorsai.track('<EVENT_NAME>', { '<EVENT_PROPERTY_KEY>': '<EVENT_PROPERTY_VALUE>', ... });
```

* Identifying an user with your own identifier. YOUR_USER_UNIQUE_IDENTIFIER can be anything which is unique among your users.
```
factorsai.identify('<YOUR_USER_UNIQUE_IDENTIFIER>');
```

* Add new user properties.
```
factorsai.addUserProperties({ '<USER_PROPERTY_KEY>': '<USER_PROPERTY_VALUE>', ... });
```