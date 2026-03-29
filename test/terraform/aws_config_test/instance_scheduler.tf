# -----------------------------------------------------------
# AWS Instance Scheduler — single-account deployment
# -----------------------------------------------------------

resource "aws_cloudformation_stack" "instance_scheduler" {
  name = "aws-instance-scheduler"

  template_url = "https://s3.amazonaws.com/solutions-reference/instance-scheduler-on-aws/latest/instance-scheduler-on-aws.template"

  capabilities = ["CAPABILITY_NAMED_IAM"]

  parameters = {
    TagName                     = "Schedule"
    DefaultTimezone             = "US/Pacific"
    SchedulingActive            = "Yes"
    CreateRdsSnapshot           = "No"
    LogRetentionDays            = "1"
    EnableSSMMaintenanceWindows = "No"
    UsingAWSOrganizations       = "No"
    RetainDataAndLogs           = "Disabled"
    OpsMonitoring               = "Disabled"
  }
}

# -----------------------------------------------------------
# A single schedule for testing — Pacific Time office hours
# -----------------------------------------------------------

resource "aws_cloudformation_stack" "instance_scheduler_schedules" {
  depends_on = [aws_cloudformation_stack.instance_scheduler]

  name = "aws-instance-scheduler-schedules"

  parameters = {
    ServiceInstanceScheduleServiceTokenARN = aws_cloudformation_stack.instance_scheduler.outputs["ServiceInstanceScheduleServiceToken"]
  }

  template_body = jsonencode({
    AWSTemplateFormatVersion = "2010-09-09"

    Parameters = {
      ServiceInstanceScheduleServiceTokenARN = {
        Type        = "String"
        Description = "Service token ARN from the Instance Scheduler stack"
      }
    }

    Resources = {
      SchedulePtOfficeHours = {
        Type = "Custom::ServiceInstanceSchedule"
        Properties = {
          ServiceToken = { Ref = "ServiceInstanceScheduleServiceTokenARN" }
          NoStackPrefix = "True"
          Name          = "schedule-pt-office-hours"
          Description   = "Pacific Time office hours (7am-5pm)"
          Timezone      = "US/Pacific"
          Periods = [
            {
              Description = "weekdays 7am-5pm Pacific Time"
              BeginTime   = "07:00"
              EndTime     = "17:00"
              WeekDays    = "mon-fri"
            }
          ]
        }
      }
    }
  })
}
