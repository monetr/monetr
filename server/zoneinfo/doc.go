// zoneinfo is a package that is meant to wrap the time.LoadLocation API.
// Instead zoneinfo.Timezone should be used instead as it takes into account
// timezone names that have changed over time. This behavior is not always
// present however and is dependent on the host operating system having timezone
// data files available that contain this information.
package zoneinfo
