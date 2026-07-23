from pyzabbix import ZabbixAPI



zapi = ZabbixAPI('http://localhost/api_jsonrpc.php')
zapi.login('Admin', 'zabbix')

# operations=[{
#             'operationtype': 1,
#             'evaltype': 0,
#             'opcommand_hst': [{'hostid': '0'}],
#             "opconditions": [
#                 {
#                     "conditiontype": 14,  # event acknowledged
#                     "operator": 0,
#                     "value": "1"
#                 }
#             ],
#             'opcommand': {
#                 'scriptid': script_id,
#             }
#         }],
#         update_operations=[{
#             'operationtype': 1,
#             # 'evaltype': 0,
#             'opcommand_hst': [{'hostid': '0'}],
#             # 'opcommand_grp': [],
#
#             # "opmessage": {
#             #     "default_msg": 0,
#             #     "message": '{TRIGGER.NAME}: {TRIGGER.STATUS}',
#             #     "subject": '{TRIGGER.NAME}: {TRIGGER.STATUS}',
#             # },
#             'opcommand': {
#                 'scriptid': script_id,
#             }
#         }]