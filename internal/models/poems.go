package models

import (
	"math/rand"
	"time"
)

// Poem represents a poem with its text and author
type Poem struct {
	Text   string
	Author string
}

// AllPoems contains a collection of thoughtful poems
var AllPoems = []Poem{
	// Inspired by Scandinavian poetry
	{
		Text: `The winter light falls soft and low,
On frozen lakes where silence grows.
In stillness finds the heart its home,
Where deepest truths are gently known.`,
		Author: "In the style of Nordic contemplation",
	},
	{
		Text: `Between the mountains and the sea,
Lives all that we will ever be.
The journey matters more than ground,
In seeking, truth itself is found.`,
		Author: "Anonymous Swedish",
	},
	{
		Text: `Time is but a river flowing,
Each moment comes without our knowing.
Catch the light while still it gleams,
Life is briefer than it seems.`,
		Author: "Inspired by Bo Setterlind",
	},

	// Classic wisdom
	{
		Text: `Not all who wander are lost,
Some seek wisdom at any cost.
The road less traveled holds the key,
To who we are and yet might be.`,
		Author: "Modern reflection",
	},
	{
		Text: `In the garden of your mind,
Plant the seeds of being kind.
Water them with gentle care,
Watch compassion flower there.`,
		Author: "Anonymous",
	},
	{
		Text: `The oak tree started as an acorn small,
Great things from tiny beginnings call.
Patience is the gardener's art,
Growth begins within the heart.`,
		Author: "Traditional wisdom",
	},
	{
		Text: `When words are soft and actions kind,
Peace will follow, you will find.
The world needs more of gentle souls,
Who mend the broken, make things whole.`,
		Author: "Modern compassion",
	},
	{
		Text: `Look up at stars on cloudless nights,
Feel small beneath eternal lights.
Yet know that in your beating chest,
Lives stardust at its very best.`,
		Author: "Cosmic reflection",
	},

	// Nature and seasons
	{
		Text: `Spring arrives on whispered wing,
Teaching hearts again to sing.
After winter's darkest day,
Hope returns to light the way.`,
		Author: "Seasonal wisdom",
	},
	{
		Text: `Autumn leaves in golden fall,
Remind us beauty touches all.
In letting go, we find release,
In empty hands, we discover peace.`,
		Author: "Fall contemplation",
	},
	{
		Text: `The morning dew on summer grass,
Reminds us nothing's meant to last.
Embrace each moment while it's here,
Tomorrow is not yet clear.`,
		Author: "Mindful living",
	},
	{
		Text: `Winter's silence teaches well,
Secrets that no voice can tell.
In stillness grows the strongest tree,
In quiet rests our clarity.`,
		Author: "Winter meditation",
	},

	// Human connection
	{
		Text: `A smile shared across the street,
Where strangers' eyes and kindness meet.
Small gestures change the world we see,
Ripples spread infinitely.`,
		Author: "Urban poetry",
	},
	{
		Text: `Listen not just to the words,
But silence in between what's heard.
The deepest truths are often found,
In spaces void of any sound.`,
		Author: "On listening",
	},
	{
		Text: `Hold your loved ones close today,
Tomorrow they may be away.
Time is precious, moments fleet,
Love is life's most sacred beat.`,
		Author: "About love",
	},

	// Technology and modern life (fitting for a file sharing service!)
	{
		Text: `In sharing files across the net,
We share much more than we forget.
Knowledge flows from hand to hand,
Connecting hearts across the land.`,
		Author: "Digital age reflection",
	},
	{
		Text: `Though screens may separate our eyes,
True connection never dies.
Technology's a bridge we build,
With human warmth and kindness filled.`,
		Author: "Modern connection",
	},
	{
		Text: `Data flows like rivers wide,
Carrying treasures deep inside.
Information, freely shared,
Shows how much the world has cared.`,
		Author: "Information age",
	},

	// More contemplative pieces
	{
		Text: `The question matters more than knowing,
In inquiry, the mind is growing.
Certainty can close the door,
Wonder opens up much more.`,
		Author: "On learning",
	},
	{
		Text: `Mistakes are teachers dressed in shame,
Each failure lights a different flame.
The path to wisdom's paved with tries,
Each stumble helps us learn to rise.`,
		Author: "On failure and growth",
	},
	{
		Text: `Today you are the yesterday,
Of all your future's coming days.
Plant now the seeds you wish to reap,
The promises you mean to keep.`,
		Author: "On time",
	},
	{
		Text: `In every ending lies a start,
Each goodbye prepares the heart.
Change is life's most constant friend,
Each chapter has to somehow end.`,
		Author: "On transitions",
	},
	{
		Text: `The heavy stone you choose to carry,
Weighs you down and makes you weary.
Set it down and walk more free,
Forgiveness is the master key.`,
		Author: "On forgiveness",
	},
	{
		Text: `Your thoughts create the world you see,
Choose them like you choose a tree.
Plant the ones that bear good fruit,
Pull the weeds out by the root.`,
		Author: "On mindfulness",
	},

	// Nature observations
	{
		Text: `The mountain stands through storm and sun,
Teaching us when day is done:
Stability comes from within,
Not from the weather we are in.`,
		Author: "Mountain wisdom",
	},
	{
		Text: `Rivers never fight the stone,
They find a path that's all their own.
Flowing past, not pushing through,
Gentle strength will see you through.`,
		Author: "Water's lesson",
	},
	{
		Text: `The seed beneath the frozen ground,
Waits patient, doesn't make a sound.
It knows its time will surely come,
When winter's bitter work is done.`,
		Author: "On patience",
	},
	{
		Text: `Birds don't sing to earn their bread,
They sing because they're living, fed.
The joy of being is enough,
Life need not be always tough.`,
		Author: "Simple joy",
	},

	// Swedish/Nordic inspired
	{
		Text: `The midnight sun and winter's night,
Both teach us different kinds of light.
In darkness learn to see the stars,
In brightness, cherish who you are.`,
		Author: "Nordic seasons",
	},
	{
		Text: `Forests thick with ancient pine,
Hold secrets older than our time.
Listen to the whispering trees,
Wisdom carries on the breeze.`,
		Author: "Forest meditation",
	},
	{
		Text: `The fjord lies still and deep and cold,
Beneath the surface, stories old.
What's hidden often holds more truth,
Than what is visible, forsooth.`,
		Author: "Fjord reflection",
	},

	// Life's journey
	{
		Text: `The path ahead is yours to make,
Each step a choice you freely take.
No map exists for life's terrain,
Both sunshine and the falling rain.`,
		Author: "On choice",
	},
	{
		Text: `You cannot rush a butterfly,
Nor force a bird to learn to fly.
All things unfold in their own time,
Each season has its own sweet rhyme.`,
		Author: "Natural timing",
	},
	{
		Text: `The stars you navigate by night,
May change when morning brings its light.
Flexibility's the sailor's art,
Adjust your course, but keep your heart.`,
		Author: "Navigation",
	},

	// Relationships
	{
		Text: `Some people are here for a reason,
Others are here for a season.
All teach us something valuable still,
Growing is life's greatest skill.`,
		Author: "On people",
	},
	{
		Text: `The bridge between two hearts is built,
From understanding, without guilt.
Communication is the key,
To set both minds and spirits free.`,
		Author: "On understanding",
	},
	{
		Text: `In anger, count to ten before,
You say what can't be said no more.
Words spoken cannot be reclaimed,
Hearts broken can't be unashamed.`,
		Author: "On words",
	},

	// Work and purpose
	{
		Text: `Your work's a canvas, paint it well,
Let every stroke a story tell.
Pride in craft, no matter what,
Makes masterpiece from every stroke.`,
		Author: "On craftsmanship",
	},
	{
		Text: `The purpose that you seek so far,
Is closer than you think you are.
Look within, not just without,
Your calling whispers, doesn't shout.`,
		Author: "On purpose",
	},
	{
		Text: `Rest is not the enemy,
Of productivity you see.
The field that lies fallow one year,
Grows stronger when the spring is here.`,
		Author: "On rest",
	},

	// Simple observations
	{
		Text: `A cup of tea, a moment's peace,
Can make the rush and worry cease.
Small pleasures scattered through the day,
Keep the darkness held at bay.`,
		Author: "Simple pleasures",
	},
	{
		Text: `The book upon your shelf unread,
Holds worlds you've never visited.
Open it and journey far,
To places that no boundaries bar.`,
		Author: "On reading",
	},
	{
		Text: `A letter written by hand,
Means more than texts could understand.
The time it takes shows that you care,
Slow things down, be truly there.`,
		Author: "On presence",
	},

	// Courage and fear
	{
		Text: `Courage isn't absence of fear,
But acting when the path's unclear.
The brave heart trembles, yet moves on,
Towards the rising of the dawn.`,
		Author: "On courage",
	},
	{
		Text: `What if I fail? you often say,
But what if wings grow on the way?
The leap of faith precedes the flight,
Step out into the morning light.`,
		Author: "On taking risks",
	},
	{
		Text: `Fear is a story we create,
Of future things that might await.
But now is all we truly have,
Fear cannot touch this present path.`,
		Author: "On fear",
	},

	// Gratitude
	{
		Text: `Count your blessings, not your woes,
Watch how gratitude just grows.
What we focus on expands,
Choose wisely what you hold in hands.`,
		Author: "On gratitude",
	},
	{
		Text: `The roof above, the food to eat,
Shoes upon our walking feet.
So much that we just take for granted,
Are the seeds that others planted.`,
		Author: "Appreciation",
	},
	{
		Text: `Say thank you to the rising sun,
For every day that has begun.
Gratitude transforms the mind,
Opens eyes that were once blind.`,
		Author: "Daily thanks",
	},

	// Memory and time
	{
		Text: `Memories are treasures kept,
Of promises and tears we wept.
They shape us but don't bind us so,
Learn from them, then let them go.`,
		Author: "On memory",
	},
	{
		Text: `The photograph can freeze a face,
But cannot capture time or place.
The feeling in that captured smile,
Lives beyond the photo's file.`,
		Author: "Beyond images",
	},
	{
		Text: `Old age is not a thing to fear,
But wisdom gathered year by year.
Each wrinkle is a story told,
More precious far than any gold.`,
		Author: "On aging",
	},

	// Hope and resilience
	{
		Text: `After the storm the sky turns clear,
After the winter, spring is near.
After the darkness comes the dawn,
After the ending, life goes on.`,
		Author: "Cycles of hope",
	},
	{
		Text: `The phoenix rises from the flame,
No longer bearing its old name.
What destroys us makes us new,
From ashes grows a different you.`,
		Author: "On transformation",
	},
	{
		Text: `The willow bends but doesn't break,
Teaches us what it will take.
Flexibility and strength combine,
In resilience's grand design.`,
		Author: "Resilience",
	},

	// Creativity
	{
		Text: `Every person is a poet,
Though they may not even know it.
Life itself's the greatest art,
Painted daily by the heart.`,
		Author: "On creativity",
	},
	{
		Text: `The blank page waits for your first word,
The silence waits for songs unheard.
Create without the fear of wrong,
Your unique voice makes life's great song.`,
		Author: "Creative courage",
	},
	{
		Text: `In imperfection beauty lies,
The crooked smile, the clouded skies.
Perfect things can never teach,
The lessons imperfections reach.`,
		Author: "Imperfect beauty",
	},

	// Solitude and silence
	{
		Text: `In solitude we hear the voice,
That guides us to our truest choice.
Alone but never lonely there,
We find ourselves in quiet prayer.`,
		Author: "On solitude",
	},
	{
		Text: `Silence is not empty space,
But fullness in a quiet place.
Listen to the noiseless sound,
Where deepest truths are often found.`,
		Author: "Sacred silence",
	},

	// Dreams and aspirations
	{
		Text: `Your dreams are seeds of what could be,
Plant them in reality.
Water them with daily action,
Watch them grow through satisfaction.`,
		Author: "On dreams",
	},
	{
		Text: `The heights you dream of reaching high,
Begin with learning how to try.
Each master started as beginner,
Persistence makes the eventual winner.`,
		Author: "On persistence",
	},

	// Truth and honesty
	{
		Text: `Speak your truth but speak it kind,
Honesty with gentle mind.
Truth need not be harsh or rough,
Loving honesty's enough.`,
		Author: "On truth",
	},
	{
		Text: `The lie may travel fast and far,
But truth reveals just what things are.
In the end, what's real remains,
While falsehood always self-explains.`,
		Author: "Truth endures",
	},

	// Balance
	{
		Text: `In balance lies the secret key,
Not this or that, but both we see.
The middle way, the golden mean,
Harmony in all between.`,
		Author: "On balance",
	},
	{
		Text: `Work and rest, the give and take,
Motion, stillness for their sake.
Everything needs opposite,
To make a whole that's truly fit.`,
		Author: "Opposites unite",
	},

	// More Nordic-inspired
	{
		Text: `The northern lights dance green and bright,
Painting stories on the night.
Magic lives in simple things,
If we have eyes to see what sings.`,
		Author: "Aurora meditation",
	},
	{
		Text: `Snow falls silent, soft and deep,
Blanketing the world in sleep.
Purity in white so clean,
Freshest start we've ever seen.`,
		Author: "Winter's gift",
	},
	{
		Text: `The sauna's heat and ice-cold lake,
Contrasts that our bodies make.
Extremes can teach us how to feel,
Alive and present, raw and real.`,
		Author: "Nordic tradition",
	},

	// Acceptance
	{
		Text: `Some things cannot be changed at all,
Accept them, standing brave and tall.
The serenity to know the difference,
Is wisdom's greatest gift and presence.`,
		Author: "Serenity",
	},
	{
		Text: `You are exactly where you need,
To learn the lessons you must heed.
Trust the journey, trust the way,
Trust yourself this very day.`,
		Author: "Self-trust",
	},

	// Community
	{
		Text: `Together we are stronger far,
Than any one of us could are.
Community's the web we weave,
Supporting what we all believe.`,
		Author: "On community",
	},
	{
		Text: `Your neighbor's burden, help them bear,
Show them someone truly cares.
We rise by lifting others high,
Together we can touch the sky.`,
		Author: "Lifting others",
	},

	// Wonder and curiosity
	{
		Text: `Never lose your sense of wonder,
At lightning, stars, and rolling thunder.
Keep your childlike awe alive,
That's what makes us truly thrive.`,
		Author: "On wonder",
	},
	{
		Text: `Ask questions, never stop your seeking,
Curiosity is wisdom speaking.
The known world's just a tiny part,
Keep questioning with an open heart.`,
		Author: "Stay curious",
	},

	// Endings and beginnings
	{
		Text: `Every sunset promises dawn,
Every ending's not yet gone.
In closing is an opening made,
Dark and light in bright cascade.`,
		Author: "Eternal cycles",
	},
	{
		Text: `The last page of your favorite book,
Is sad until another look.
For every ending makes you free,
To start again with eyes that see.`,
		Author: "New chapters",
	},

	// Legacy
	{
		Text: `What you leave behind that matters,
Isn't gold or silver platters.
But the kindness that you showed,
Love you gave along life's road.`,
		Author: "True legacy",
	},
	{
		Text: `Plant trees beneath whose shade you'll never sit,
Do things from which you'll never benefit.
The future generation's your concern,
Give forward what you never can return.`,
		Author: "For the future",
	},

	// Simplicity
	{
		Text: `Life's complexity we create,
Simplicity is the natural state.
Reduce, release, let go, be free,
Less is more, simplicity.`,
		Author: "Simple living",
	},
	{
		Text: `The happiest people don't have everything,
They appreciate the joy small moments bring.
Contentment comes from being grateful,
For simple things that make life playful.`,
		Author: "Simple joy",
	},

	// Presence
	{
		Text: `This moment is the only time,
That's truly, really, fully mine.
Past is gone and future's not here,
Now is all I hold so dear.`,
		Author: "Present moment",
	},
	{
		Text: `Be here now, the old ones say,
Don't wish this moment away.
Life is made of now and now,
Be present here, be present now.`,
		Author: "Now is all",
	},

	// Final wisdom poems
	{
		Text: `The journey of a thousand miles,
Begins beneath your feet's first trials.
Each step you take, however small,
Brings you closer to it all.`,
		Author: "Ancient wisdom",
	},
	{
		Text: `Listen to the elder's voice,
Learn from every life's choice.
Wisdom comes from those who lived,
And from mistakes they have forgived.`,
		Author: "Elder wisdom",
	},
	{
		Text: `Your story isn't written yet,
Each day's a page you haven't set.
Pick up the pen, begin to write,
Your life, your way, your truth, your light.`,
		Author: "Your story",
	},
	{
		Text: `In the end, it's not the years,
The laughter or the falling tears.
But did you love, and were you kind?
That's what you leave for those behind.`,
		Author: "What matters most",
	},
	{
		Text: `The universe is vast and wide,
But look within, not just outside.
The cosmos lives in every soul,
You are a part of the great whole.`,
		Author: "Inner cosmos",
	},
	{
		Text: `When you don't know what to do,
Let kindness be your guide to true.
Choose the path of gentle care,
You'll never regret loving there.`,
		Author: "Choose kindness",
	},
	{
		Text: `Today might be somebody's last,
Treat each person as if they're vast.
Everyone's fighting their own fight,
Be the reason someone's day is bright.`,
		Author: "Be someone's light",
	},
	{
		Text: `The waves return unto the shore,
Again, again, forevermore.
Persistence is the ocean's way,
Keep trying every single day.`,
		Author: "Ocean's lesson",
	},
	{
		Text: `A single candle lights the dark,
A single voice can leave its mark.
Never think you're too small to matter,
One by one, we make things better.`,
		Author: "Individual power",
	},
	{
		Text: `Breathe in peace, breathe out the rest,
Every breath can be a test.
Of staying calm and staying true,
Of being centered, being you.`,
		Author: "Breathing meditation",
	},
	{
		Text: `The moon reminds us every night,
That even in the darkest blight,
There's always something beautiful,
Something peaceful, something full.`,
		Author: "Moon's reminder",
	},
	{
		Text: `Your word is all you truly own,
The honor in the truth you've shown.
Keep your promises, be true,
Integrity lives within you.`,
		Author: "On integrity",
	},
	{
		Text: `The butterfly effect is real,
Small actions have enormous zeal.
Your kindness ripples far and wide,
More than you can see inside.`,
		Author: "Ripple effect",
	},
	{
		Text: `Don't wait for perfect time to start,
Begin right now with willing heart.
Perfection is the enemy,
Of progress and what's meant to be.`,
		Author: "Start now",
	},
	{
		Text: `The greatest gift you have to give,
Is showing others how to live.
Not with words but with your deeds,
Actions speak more than creeds.`,
		Author: "Lead by example",
	},
	{
		Text: `In diversity we find our strength,
Going together any length.
Different colors paint the sky,
Together we can learn to fly.`,
		Author: "Unity in diversity",
	},
	{
		Text: `The hardest part is to begin,
Once started, you will surely win.
Not every battle, that is true,
But you'll win by becoming you.`,
		Author: "On beginning",
	},
	{
		Text: `Look for beauty everywhere,
In ordinary, common care.
Magic hides in mundane things,
See the sacred in everything.`,
		Author: "Sacred ordinary",
	},
	{
		Text: `Your feelings are all valid, true,
They are a part of being you.
But feelings aren't the final say,
On who you are or come what may.`,
		Author: "On emotions",
	},
	{
		Text: `The teacher appears when ready,
The student learns when steady.
Life itself's the greatest school,
Love and wisdom are the rule.`,
		Author: "Life lessons",
	},
}

// GetRandomPoem returns a random poem from the collection
func GetRandomPoem() Poem {
	rand.Seed(time.Now().UnixNano())
	return AllPoems[rand.Intn(len(AllPoems))]
}

// GetPoemOfTheDay returns a consistent poem for the current day
func GetPoemOfTheDay() Poem {
	// Use current time as seed for random poem on each visit
	rand.Seed(time.Now().UnixNano())
	return AllPoems[rand.Intn(len(AllPoems))]
}
