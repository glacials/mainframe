#!/usr/bin/env ruby

# Run this script at boot to continuously take photos, put them in S3, and
# occasionally collate them into a timelapse video which also goes into S3.

require "fileutils"

timelapse = "treefort"
HOME = "/home/glacials/s3/timelapse/#{timelapse}".freeze

# Set up mount point and directory tree
`sleep 10` # Can't do it right away though

`s3fs glacials-owl ~/s3 -o allow_other -o umask=0007,uid=1001,gid=1001`
`mkdir -p ~/s3/ops ~/s3/timelapse/#{timelapse}/photos ~/s3/timelapse/#{timelapse}/videos`

def syncrun(prefix, cmd)
  puts "#{Time.now.utc.strftime("%FT%TZ")} [#{prefix}] #{cmd}"
  result = `#{cmd}`
  puts result if result
  puts "#{Time.now.utc.strftime("%FT%TZ")} [#{prefix}] Done."
end

def asyncrun(prefix, cmd)
  pid = spawn(cmd)
  Process.detach(pid)
  puts "#{Time.now.utc.strftime("%FT%TZ")} [#{prefix}] Forked."
end

def takephoto
  now = Time.now.utc
  filename = "#{now.strftime("%FT%TZ")}.jpg"

  syncrun("photo", "raspistill --output - --nopreview > #{HOME}/photos/#{filename}")
end

def setinternalip
  asyncrun("internalip", "ip address > ~/s3/ops/ip.txt")
end

def setexternalip
  asyncrun("externalip", "curl --silent ifconfig.io > ~/s3/ops/external-ip.txt")
end

def makevideo(force: false)
  prefix = "video"

  now = Time.now.utc
  if now.hour > 1 # Only start running between 00:00 - 02:00
    unless force
      puts "[#{prefix}] Not rendering video because not in the right hours."
      return
    end
    puts "#{Time.now.utc.strftime("%FT%TZ")} [#{prefix}] Forcing video render even outside allowed hours."
  end

  tmpfile = "/tmp/#{now.strftime("%FT%TZ")}.mp4"
  prefix = "video"

  # Remove blank (corrupted) files so ffmpeg doesn't trip on them, and remove
  # small (nighttime, all black) photos to make the timelapse not be boring
  cmd = "find #{HOME}/photos -type f -size -1M -delete"
  cmd += " && "
  # Add -vf rotate=-90*PI/180 before #{tmpfile} to rotate 90 degrees
  cmd += "ffmpeg -y -framerate 30 -pattern_type glob -i \"#{HOME}/photos/*.jpg\" -s:v 1440x1080 -pix_fmt yuv420p -c:v libx264 -crf 17 -pix_fmt yuv422p #{tmpfile} && mv #{tmpfile} #{HOME}/videos"

  puts "#{Time.now.utc.strftime("%FT%TZ")} [#{prefix}] Starting prune, render, and upload."
  puts "#{Time.now.utc.strftime("%FT%TZ")} [#{prefix}] #{cmd}"

  asyncrun(prefix, cmd)
end

ACTIONS = [
  {
    name: "internalip",
    interval: 300,
    lambda: -> { setinternalip }
  },
  {
    name: "externalip",
    interval: 300,
    lambda: -> { setexternalip }
  },
  {
    name: "photo",
    interval: 1,
    lambda: -> { takephoto }
  },
  {
    name: "video",
    # interval: 60*60, # ATTEMPT to run 1/hour, but makevideo func restricts to nighttime
    interval: -1, # Never run; for medium+ lengths it never makes it through before crashing
    lambda: -> { makevideo(force: false) } # If you force, also increase interval to not get overlap
  }
]

# Start just before an interval that you want to do on boot (set IPs), but far
# away from one you don't (render video). Or for dev, do both.
START = 60 * 60 - 5

def main
  max = ACTIONS.map { |a| a[:interval] }.max
  i = START

  Kernel.loop do
    ACTIONS.each do |action|
      next if action[:interval].negative?

      action[:lambda].call if i % action[:interval].zero?
    end

    sleep 1

    i += 1
    i = 0 if i >= max
  end
end

main
