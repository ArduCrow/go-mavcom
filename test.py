import pymavlink
import pymavlink.mavutil

# connect to vehicle and get target id, system id, and component id
vehicle = pymavlink.mavutil.mavlink_connection("127.0.0.1:14550")
target_id = vehicle.target_system
print(vehicle)
print(vehicle.target_system)

# get heartbeat
vehicle.wait_heartbeat()
print(vehicle.messages)
# print heartbeat data
print(vehicle.messages['HEARTBEAT'])

# send takeoff command
vehicle.mav.command_long_send(
    vehicle.target_system, vehicle.target_component,
    pymavlink.mavutil.mavlink.MAV_CMD_NAV_TAKEOFF,
    0, 0, 0, 0, 0, 0, 0, 0)