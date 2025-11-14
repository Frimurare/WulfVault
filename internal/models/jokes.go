// WulfVault - Secure File Transfer System
// Copyright (c) 2025 Ulf Holmström (Frimurare)
// Licensed under the GNU Affero General Public License v3.0 (AGPL-3.0)
// You must retain this notice in any copy or derivative work.

package models

import (
	"math/rand"
	"time"
)

// FileJoke represents a one-liner joke about file sharing
type FileJoke struct {
	Text string
}

// AllJokes contains a collection of one-liner jokes about file sharing
var AllJokes = []FileJoke{
	{Text: "My email said '25MB attachment limit exceeded.' That's when I knew we needed a better solution."},
	{Text: "I tried emailing a video file once. My inbox is still apologizing to the mail server."},
	{Text: "Email attachments: Where gigabytes go to get rejected."},
	{Text: "The file is too large for email? Welcome to WulfVault, where size doesn't matter."},
	{Text: "I don't always send large files, but when I do, I don't use email."},
	{Text: "Email said 'file too large.' I said 'hold my beer' and opened WulfVault."},
	{Text: "Why did the video file cross the internet? Because email wouldn't let it through."},
	{Text: "My 4K video tried to fit through email. It's still stuck at the entrance."},
	{Text: "Email: 'Your file is too large.' Me: 'Your limit is too small.'"},
	{Text: "I sent a 2GB file via email once. Just kidding, nobody has ever done that successfully."},
	{Text: "The cloud called. It wants its storage space back from your email server."},
	{Text: "Email attachments are like airplane seats - they keep getting smaller while we keep getting bigger."},
	{Text: "My CAD files laughed at the 25MB email limit. Then I showed them WulfVault."},
	{Text: "If your file is too large for email, you're not using the wrong file - you're using the wrong tool."},
	{Text: "Email was invented before high-resolution existed. It shows."},
	{Text: "I tried to email a raw photo collection. The mail server filed for therapy."},
	{Text: "25MB email limit: Because it's still 1995 somewhere in the email infrastructure."},
	{Text: "The file was 500MB. The email server said no. WulfVault said 'go ahead, make my day.'"},
	{Text: "Why compress your files when you can just use proper file sharing?"},
	{Text: "Email rejected my video. WulfVault gave it a red carpet welcome."},
	{Text: "I uploaded a 3GB file. Email would have needed a week. WulfVault needed a minute."},
	{Text: "My ZIP file was so compressed it needed decompression therapy. Then I found WulfVault."},
	{Text: "Email limits: Making life harder since... well, forever."},
	{Text: "The best part of email attachments? The error messages. Wait, I mean WulfVault."},
	{Text: "I asked email to send my 1GB file. It's still laughing at me."},
	{Text: "File sharing in 2025: Either you evolve or you email 47 parts of a ZIP file."},
	{Text: "My design portfolio was too large for email. Good thing professionalism isn't."},
	{Text: "Email: 'File too large.' WulfVault: 'File just right.'"},
	{Text: "Why did I stop using email attachments? Because I value my sanity."},
	{Text: "The only thing slower than emailing large files is... actually, nothing is slower."},
	{Text: "I tried to email a database backup. The IT department is still recovering."},
	{Text: "WulfVault: Where your files go when email breaks up with them."},
	{Text: "Email said my file was too large. I said their vision was too small."},
	{Text: "My project files asked email for a ride. Email said 'I don't do carpools that big.'"},
	{Text: "The 25MB limit isn't a feature, it's a fossil."},
	{Text: "I don't compress files anymore. I just use WulfVault like a civilized person."},
	{Text: "Email attachments: Where good files go to get bounced."},
	{Text: "My video file was so large, email thought it was a prank."},
	{Text: "Why split files when you can just... not?"},
	{Text: "Email rejected my presentation. Apparently 300MB of charts is 'too much information.'"},
	{Text: "I tried emailing a VM image. The mail server immediately called for backup."},
	{Text: "WulfVault: Because 'file too large' should never be in your vocabulary."},
	{Text: "Email limits are like speed limits from the horse-and-buggy era."},
	{Text: "My 4K drone footage laughed at the email attachment limit. Then it flew to WulfVault."},
	{Text: "The file was 800MB. Email said no. I said WulfVault said yes."},
	{Text: "Why do email servers fear large files? Because they remember the 90s."},
	{Text: "I sent a large file via email once. The 'sending...' progress bar is still going."},
	{Text: "Email: Built for text. Used for everything. Succeeds at nothing over 25MB."},
	{Text: "My raw footage wanted to email itself. The therapist said it was self-destructive behavior."},
	{Text: "WulfVault handles large files like email handles small text - perfectly."},
	{Text: "Email server: 'File too large.' Me: 'Server too small.'"},
	{Text: "I compressed my files until they cried. Then I discovered WulfVault and gave them dignity."},
	{Text: "The only thing email attachments taught me is that there's a better way."},
	{Text: "My backup archive tried to email itself. It's still in therapy."},
	{Text: "WulfVault: Where files don't need to lose weight to travel."},
	{Text: "Email said my file was obese. WulfVault said my file was beautiful just the way it is."},
	{Text: "I tried to email a Blender project. The mail server filed for early retirement."},
	{Text: "Why do we still have 25MB limits? Because some people enjoy suffering."},
	{Text: "My podcast episode was too large for email. Good thing my audience uses WulfVault."},
	{Text: "Email: 'Cannot send file.' WulfVault: 'Hold my beer and watch this.'"},
	{Text: "The file was 2GB. Email panicked. WulfVault yawned."},
	{Text: "I don't split ZIP files anymore. I have WulfVault and self-respect."},
	{Text: "Email rejected my video. I rejected email. We're both happier now."},
	{Text: "WulfVault: Because your files deserve first-class treatment, not email's baggage restrictions."},
	{Text: "My architectural plans were too detailed for email. Email prefers stick figures apparently."},
	{Text: "The mail server saw my 500MB attachment and immediately updated its resume."},
	{Text: "Email limits: Keeping file sharing in the stone age since the bronze age."},
	{Text: "I asked email to send my video. It's still buffering the question."},
	{Text: "WulfVault handles terabytes. Email handles tears of frustration."},
	{Text: "My music production files tried email. Email tried therapy."},
	{Text: "Why did the large file avoid email? Because it had standards."},
	{Text: "Email: Built for memos. Used for media. Fails at files."},
	{Text: "I tried emailing a game build. The mail server rage-quit."},
	{Text: "WulfVault: Where 'too large' is just another way of saying 'just right.'"},
	{Text: "My raw camera footage looked at email limits and started laughing."},
	{Text: "Email said 'file too large.' I said 'expectations too low.'"},
	{Text: "The only thing email is good at rejecting is large files and my self-esteem."},
	{Text: "I compressed my files 10 times. Or I could just use WulfVault once."},
	{Text: "Why do email servers hate large files? Professional jealousy."},
	{Text: "My project backup was too big for email. Email was too small for my project."},
	{Text: "WulfVault: Because splitting files into 47 parts is not a solution."},
	{Text: "Email said my ISO was too large. I said their vision was too small."},
	{Text: "I tried to email a database dump. The database dumped email instead."},
	{Text: "The mail server saw my 1GB file and immediately called in sick."},
	{Text: "WulfVault handles what email can only dream about."},
	{Text: "My design files were too heavy for email. Email needs to hit the gym."},
	{Text: "Why compress creativity when you can just transfer it properly?"},
	{Text: "Email: Where large files go to get body-shamed."},
	{Text: "I don't lose quality for email limits anymore. I have standards and WulfVault."},
	{Text: "My video project laughed at email. Email cried. I used WulfVault."},
	{Text: "The file was perfect. Email wanted it smaller. WulfVault wanted it as-is."},
	{Text: "Why did I switch from email? Because 'attachment_part_23.zip' isn't professional."},
	{Text: "Email limits are like diet culture for files - toxic and outdated."},
	{Text: "I tried emailing a VM. The mail server is still loading."},
	{Text: "WulfVault: Where files don't need to fit into email's skinny jeans."},
	{Text: "My 3D renders were too beautiful for email. Email prefers low-poly apparently."},
	{Text: "The mail server saw my footage and immediately checked its retirement fund."},
	{Text: "Email rejected my files. WulfVault accepted me for who I am."},
	{Text: "Why split one file into 20 when you can send one file as one?"},
	{Text: "I asked email to handle my backup. Email asked for a backup plan."},
	{Text: "WulfVault: Because your files shouldn't need to diet to travel."},
	{Text: "My source code with assets was too large. Email said 'git gud.' I said 'WulfVault better.'"},
	{Text: "The only limit at WulfVault is your imagination. And maybe disk space. But definitely not email size."},
	{Text: "Email saw my file size and experienced an existential crisis."},
	{Text: "I tried to email a recording session. The mail server went into witness protection."},
	{Text: "WulfVault: Where 'file too large' is a phrase you'll forget exists."},
	{Text: "My drone footage was too epic for email. Email prefers mediocrity."},
	{Text: "Why do we tolerate email limits? Because we didn't know about WulfVault yet."},
	{Text: "Email said no. WulfVault said yes. I said goodbye email."},
	{Text: "The file was massive. Email was scared. WulfVault was ready."},
	{Text: "I don't split files anymore. I'm not a file surgeon."},
	{Text: "My photo library tried email. Email tried quitting."},
	{Text: "WulfVault: Because professional work requires professional tools."},
	{Text: "Email limits: The participation trophy of file transfer."},
	{Text: "I tried emailing my graduation video. It's still trying to graduate from the outbox."},
	{Text: "Why compress quality when you can just use quality file sharing?"},
	{Text: "Email rejected my CAD file. The drawing was technical. The rejection was personal."},
	{Text: "My backup was 5GB. Email fainted. WulfVault shrugged."},
	{Text: "WulfVault: Where your files travel business class, not email economy."},
	{Text: "The mail server saw my project file and immediately updated its will."},
	{Text: "Email is great for text. WulfVault is great for everything else."},
	{Text: "I tried to email a conference recording. The server is still conferencing about it."},
	{Text: "Why did the large file join WulfVault? Because email broke its heart too many times."},
	{Text: "My renders were too realistic for email. Email prefers fantasy like 'this will send successfully.'"},
	{Text: "Email said 'file too large.' I said 'tool too old.'"},
	{Text: "The only thing email compresses well is my patience."},
	{Text: "I tried emailing a virtual machine. It's still virtualizing the attempt."},
	{Text: "WulfVault: Because your files deserve respect, not rejection."},
	{Text: "My master thesis was too long for email. Email apparently has attention span issues."},
}

// GetJokeOfTheDay returns a random joke that changes every 5 seconds
func GetJokeOfTheDay() FileJoke {
	// Use 5-second intervals for stable joke display
	fiveSecondInterval := time.Now().Unix() / 5
	rand.Seed(fiveSecondInterval)
	return AllJokes[rand.Intn(len(AllJokes))]
}

// GetRandomJoke returns a random joke from the collection
func GetRandomJoke() FileJoke {
	rand.Seed(time.Now().Unix() / 5)
	return AllJokes[rand.Intn(len(AllJokes))]
}
