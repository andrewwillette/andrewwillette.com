package server

import (
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

type DropboxSheetMusic struct {
	Name string
	URL  string
}

type Sheets []DropboxSheetMusic

type SheetMusicPageData struct {
	Sheets      Sheets
	CurrentYear int
}

var (
	sheetmusicData = SheetMusicPageData{
		Sheets: []DropboxSheetMusic{
			{
				Name: "Indian Killed a Woodcock",
				URL:  "https://www.dropbox.com/scl/fi/d25emkkpmnia26slgwsi5/Indian-Killed-a-Woodcock.pdf?rlkey=u00g0oqbs49axngbjlu897a42&st=qy9phemg&dl=0",
			},
			{
				Name: "Sweet Bunch of Daisies",
				URL:  "https://www.dropbox.com/scl/fi/7hbls71as3ejvfwex3ntt/Sweet-Bunch-of-Daisies.pdf?rlkey=3tblcbmpkw3kz5g0rpmyqp7hx&st=wqk82xth&dl=0",
			},
			{
				Name: "Brandywine",
				URL:  "https://www.dropbox.com/scl/fi/31t8fw0n15zuimfp20hfz/Brandywine.pdf?rlkey=k209e9ffnlg905uyjuddl92it&st=b5beln9l&dl=0",
			},
			{
				Name: "Jerusalem Ridge",
				URL:  "https://www.dropbox.com/scl/fi/dl29q5axt6senm6vdu17b/Jerusalem-Ridge.pdf?rlkey=q71yvhbsx0uk2497rku2c5a5m&st=akxxqhhs&dl=0",
			},
			{
				Name: "Salt Spring",
				URL:  "https://www.dropbox.com/scl/fi/fbfi8h5smc80jl33jnxq7/Salt-Spring.pdf?rlkey=0g1oua5tmne6nnpc2a77pxr9v&st=y8aa5oln&dl=0",
			},
			{
				Name: "Washington Country",
				URL:  "https://www.dropbox.com/scl/fi/m7nk6d184ogwstp96q8y2/Washington-County.pdf?rlkey=wpdk68m4r0l2wn3a29ia0qwv3&st=bw0bz5ia&dl=0",
			},
			{
				Name: "Lonesome Moonlight Waltz",
				URL:  "https://www.dropbox.com/scl/fi/mivxvgxs2mtf3lv5syiq9/Lonesome-Moonlight-Waltz.pdf?rlkey=5o2w92cjqq5yimeqbyn38n2l1&st=gl6janvk&dl=0",
			},
			{
				Name: "Come All Ye Fair and Tender Ladies",
				URL:  "https://www.dropbox.com/scl/fi/y0y9bfb0sfq1ri95gg3es/Come-All-Ye-Fair-and-Tender-Ladies.pdf?rlkey=evyk6evqhrr5dgf6cr56rln46&st=8ep67jcp&dl=0",
			},
			{
				Name: "Pass Me Not",
				URL:  "https://www.dropbox.com/scl/fi/l1npmslthn2hu0xpt7jmj/Pass-Me-Not.pdf?rlkey=q3x1tr5ih8sehevyfg3xu7dwp&st=5wtbjcd3&dl=0",
			},
			{
				Name: "Bluegrass in the Backwoods",
				URL:  "https://www.dropbox.com/s/2ihcmsfff560qdn/Bluegrass%20in%20the%20Backwoods.pdf?st=kf7g4fvn&dl=0",
			},
			{
				Name: "Huckleberry Hornpipe",
				URL:  "https://www.dropbox.com/scl/fi/8p594n4u9ag78n3ozswy5/Huckleberry-Hornpipe.pdf?rlkey=m6hc9lwjh51fur3a9os5dx9cw&st=cp9i8u0t&dl=0",
			},
			{
				Name: "Angel's Waltz",
				URL:  "https://www.dropbox.com/scl/fi/mggzgy6r0mlneufybz55v/Angel-s-Waltz.pdf?rlkey=po24evmirkr7596fz655k961h&st=e4gmdr70&dl=0",
			},
			{
				Name: "Cumberland Gap",
				URL:  "https://www.dropbox.com/scl/fi/9vnjhsojyefsutz4yzt00/Cumberland-Gap.pdf?rlkey=i8ueptsmvhfmi59ww7h3q9dij&dl=0",
			},
			{
				Name: "Texas Crapshooter",
				URL:  "https://www.dropbox.com/scl/fi/3a6rtp1nr9gsp7v1o96ue/Texas-Crapshooter.pdf?rlkey=e7ofjv00sirynyj30j5lkdalg&dl=0",
			},
			{
				Name: "Benton's Dream",
				URL:  "https://www.dropbox.com/scl/fi/i4c0x7z8i8eis0gvyxqrl/Benton-s-Dream.pdf?dl=0&rlkey=ra3i5gf5kyu6ulup5uqezpzr9",
			},
			{
				Name: "Tangle Weed",
				URL:  "https://www.dropbox.com/scl/fi/evg8lz6gtllvk3hxodw19/Tangle-Weed.pdf?rlkey=dwpxnc9q3trzumo6ay01roc91&st=qnpwtith&dl=0",
			},
			{
				Name: "Cherokee Shuffle",
				URL:  "https://www.dropbox.com/scl/fi/lymfxdun2j66pmdyrcvr7/Cherokee-Shuffle.pdf?rlkey=9p1tkuvnv2i84tj1gswaqh8r0&st=tgfnuhou&dl=0",
			},
			{
				Name: "Polecat Blues",
				URL:  "https://www.dropbox.com/scl/fi/vbujvvahgnm2x42r8ynls/Polecat-Blues.pdf?rlkey=4kdf38nmotl55mkdbk915xm86&dl=0",
			},
			{
				Name: "Sherry",
				URL:  "https://www.dropbox.com/scl/fi/6cw6uzemja7knvau99n9r/Sherry.pdf?rlkey=u0qpuglnqj0j499e4uojxo0sb&dl=0",
			},
			{
				Name: "Leather Britches",
				URL:  "https://www.dropbox.com/scl/fi/pn4pnf797x0iqaonr1a8h/Leather-Britches.pdf?rlkey=lvzzvsank7s58vhhk9mpuw2gh&dl=0",
			},
			{
				Name: "Carrol County Blues",
				URL:  "https://www.dropbox.com/scl/fi/67b37hd50xn69nqr37o32/Carrol-county-blues.pdf?rlkey=tmrcqs2qvpaaatgrtvc74yeor&dl=0",
			},
			{
				Name: "Big Mon",
				URL:  "https://www.dropbox.com/scl/fi/hqufi023ghk18v6f3ww5q/Big-Mon.pdf?rlkey=aly0ph2tjf671oz4n4glswena&dl=0",
			},
			{
				Name: "Flop Eared Mule",
				URL:  "https://www.dropbox.com/scl/fi/6gftl0pl4qwmal7nakjt6/Flop-Eared-Mule.pdf?rlkey=aeozb6gwvgdlg1813c8shzqc4&dl=0",
			},
			{
				Name: "Down South In Dixie",
				URL:  "https://www.dropbox.com/scl/fi/5dvldnguxxmn3vy5n8yzm/Down-South-In-Dixie.pdf?rlkey=1nvtabqzzbvf3pm1th6rrafvz&dl=0",
			},
			{
				Name: "Wednesday Night Waltz",
				URL:  "https://www.dropbox.com/scl/fi/vri8tdzgjddy6bzugq2ma/Wednesday-Night-Waltz.pdf?rlkey=jx2hjze0nqfa2pmvze7w8pjpi&dl=0",
			},
			{
				Name: "Down The Road",
				URL:  "https://www.dropbox.com/scl/fi/ioz5a40t97cohit78uvgi/Theres-A-Brownskin-Girl-Down-The-Road-Somewhere.pdf?rlkey=8nzc2p6a25oi4qky4hm5n8gln&dl=0",
			},
			{
				Name: "Sugar In The Gourd",
				URL:  "https://www.dropbox.com/scl/fi/pixivlknf2lq9rm1cq3gs/Sugar-in-the-Gourd.pdf?rlkey=a049bgr8vghi7qwr5cck3udd7&dl=0",
			},
			{
				Name: "Walkin' In My Sleep",
				URL:  "https://www.dropbox.com/scl/fi/jf0spl77gw15hk5c99t4r/Walking-In-My-Sleep.pdf?rlkey=jfx1jp44l3c2cr8vqwohcffw3&dl=0",
			},
		},
	}
)

// Len to implement sort.Interface
func (sheets Sheets) Len() int {
	return len(sheets)
}

// Swap to implement sort.Interface
func (sheets Sheets) Swap(i, j int) {
	sheets[i], sheets[j] = sheets[j], sheets[i]
}

// handleSheetmusicPage handles returning the transcription template
func handleSheetmusicPage(c echo.Context) error {
	sort.Sort(sheetmusicData.Sheets)
	sheetmusicData.CurrentYear = time.Now().Year()
	err := c.Render(http.StatusOK, "sheetmusicpage", sheetmusicData)
	if err != nil {
		return err
	}
	return nil
}

// Less to implement sort.Interface
func (sheets Sheets) Less(i, j int) bool {
	switch strings.Compare(sheets[i].Name, sheets[j].Name) {
	case -1:
		return true
	case 0:
		return false
	case 1:
		return false
	default:
		return false
	}
}
