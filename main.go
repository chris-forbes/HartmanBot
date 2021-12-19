package main

import (
	"flag"
	"fmt"
	"image/color"
	"os"
	"os/signal"
	"regexp"
	image "samm-bot/model"
	service "samm-bot/services"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	embed "github.com/clinet/discordgo-embed"
	"github.com/fogleman/gg"
)

// I think this is going to be out command line token?
var (
	Token string
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot token")
	flag.Parse()
}

func main() {
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("Somethings gone badly wrong, ", err)
		return
	}

	dg.AddHandler(messageCreate)
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	err = dg.Open()
	if err != nil {
		fmt.Println("Failed to create connection, ", err)
		return
	}

	fmt.Println("Connection to discord successful, bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dg.Close()
}

func overlayTextOnImage(request image.ImageRequest, s *discordgo.Session, m *discordgo.MessageCreate) {
	fmt.Println("Overlaying text on image")
	// Load the background image
	bgImage, err := gg.LoadPNG(request.BgImgPath)
	if err != nil {
		fmt.Println("Failed to load background image, ", err)
		return
	}
	imgWidth := bgImage.Bounds().Dx()
	imgHeight := bgImage.Bounds().Dy()

	dc := gg.NewContext(imgWidth, imgHeight)
	dc.DrawImage(bgImage, 0, 0)

	if err := dc.LoadFontFace(request.FontPath, request.FontSize); err != nil {
		fmt.Println("Failed to load font, ", err)
		return
	}

	x := float64(imgWidth / 2)
	y := float64((imgHeight / 2) + 150)

minimizeY: // label for the goto statement
	if y > float64(imgHeight) {
		y -= 10
		goto minimizeY
	}

	maxWidth := float64(imgWidth) - 60.0

	dc.SetColor(color.White)

	dc.DrawStringWrapped(request.Text, x, y, 0.5, 0.5, maxWidth, 1.5, gg.AlignCenter)

	dc.SavePNG("resources/temp.png")
	sendMessageWithLocalFileResource(s, m, "", "resources/temp.png")
}

func sendMessageWithLocalFileResource(s *discordgo.Session, m *discordgo.MessageCreate, content string, localFile string) {
	// by this point assume the message content is ready to go
	// load our file and work out its size until the point we're ready to read the file
	file, err := os.Open(localFile)
	if err != nil {
		fmt.Println("Failed to open file,", err)
		return
	}

	defer file.Close()

	_, err = s.ChannelFileSendWithMessage(m.ChannelID, content, "sarge.png", file)
	if err != nil {
		fmt.Println("Failed to send file message,", err)
		return
	}
}

func sendEmbedWithImageFileURL(s *discordgo.Session, m *discordgo.MessageCreate, content string, imageUrl string, embedTitle string) {
	embedFrame := embed.NewEmbed()
	embedFrame.SetURL(imageUrl)
	embedFrame.SetTitle(embedTitle)
	embedFrame.SetDescription(content)
	_, err := s.ChannelMessageSendEmbed(m.ChannelID, embedFrame.MessageEmbed)
	if err != nil {
		fmt.Println("Failed to send embed, ", err)
		return
	}
}

func sendSergentMesage(s *discordgo.Session, content string, m *discordgo.MessageCreate) {
	request := image.ImageRequest{
		BgImgPath: "resources/Sgt_Hartman.png",
		FontPath:  "resources/JuliaMono-ExtraBold.ttf",
		FontSize:  float64(50.0) - float64(len(content)/2),
		Text:      content,
	}
	overlayTextOnImage(request, s, m)
}

func WordCount(value string) int {
	// Match non-space character sequences.
	re := regexp.MustCompile(`[\S]+`)

	// Find all matches and return count.
	results := re.FindAllString(value, -1)
	return len(results)
}

func messageCreate(session *discordgo.Session, discordMessage *discordgo.MessageCreate) {
	//Don't answer our own messages
	if discordMessage.Author.ID == session.State.User.ID {
		return
	}

	if strings.HasPrefix(discordMessage.Content, "!help") {
		help_message := `
> Gunnery Sgt. Hartman Bot 
> version: 0.0.3-pre-alpha
> Commands:		
> 	!help - This message
> 	!sarge <text> - Sarge the bot with the text
> 	!meme - receive a random meme`
		session.ChannelMessageSend(discordMessage.ChannelID, help_message)
	}

	if strings.HasPrefix(discordMessage.Content, "!meme") {
		randomMeme := service.GetRandomMeme()
		_, err := session.ChannelMessageSend(discordMessage.ChannelID, randomMeme)
		if err != nil {
			fmt.Println("Failed to send message to channel ", discordMessage.ChannelID)
			fmt.Println("Error message, ", err)
		}
	}

	if strings.HasPrefix(discordMessage.Content, "!sargent") || strings.HasPrefix(discordMessage.Content, "!sarge") || strings.HasPrefix(discordMessage.Content, "!sergeant") || strings.HasPrefix(discordMessage.Content, "!serge") {
		body := strings.SplitN(discordMessage.Content, " ", 2)
		if len(body) == 2 {
			fmt.Println("Wordcount is: ", WordCount(body[1]))
			if WordCount(body[1]) > 10 {
				sendSergentMesage(session, "10 words or less maggot", discordMessage)
			} else {
				sendSergentMesage(session, body[1], discordMessage)
			}
		} else {
			fmt.Println("Failed to parse body properly, ", body)
		}
	}

	if discordMessage.MentionEveryone {
		session.ChannelMessageDelete(discordMessage.ChannelID, discordMessage.Reference().MessageID)
		responseMessage := fmt.Sprintf(`
Who just pinged @everyone? Who the fuck did that? Who's the slimy little communist shit twinkle-toed cocksucker down here, who just signed his own death warrant? 
Nobody, huh? The fairy fucking godmother pinged @everyone ! Out-fuckingstanding! I will P.T. you all until you fucking die! 
I'll P.T. you until your assholes are sucking buttermilk.

Whats that twinkle-toes you wanted to tell @everyone: 

%s`, discordMessage.Content)
		sendMessageWithLocalFileResource(session, discordMessage, responseMessage, "resources/Sgt_Hartman.png")
	}
}
