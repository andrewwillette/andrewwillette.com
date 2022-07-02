package persistence

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type SoundcloudTestSuite struct {
	suite.Suite
}

func TestSoundcloudSuite(t *testing.T) {
	suite.Run(t, new(SoundcloudTestSuite))
}

func (suite *SoundcloudTestSuite) SetupTest() {
	deleteTestDatabase()
}

func (suite *SoundcloudTestSuite) TearDownSuite() {
	deleteTestDatabase()
}

func (suite *SoundcloudTestSuite) TestSoundcloudUrlService() {
	soundcloudUrlService := &SoundcloudUrlService{SqliteFile: testDatabaseFile}
	soundcloudUrlService.createSoundcloudUrlTable()
	tables, err := getAllTables(testDatabaseFile)
	if err != nil {
		suite.T().Fail()
	}
	assert.Contains(suite.T(), tables, soundcloudTable)
	soundcloudUrlOne := SoundcloudUrl{Url: "soundcloud.com/example", UiOrder: 0}
	err = soundcloudUrlService.AddSoundcloudUrl(soundcloudUrlOne.Url)
	if err != nil {
		suite.T().Fail()
	}
	soundcloudUrls, err := soundcloudUrlService.GetAllSoundcloudUrls()
	if err != nil {
		suite.T().Fail()
	}
	assert.True(suite.T(), soundcloudUrlExists(soundcloudUrls, soundcloudUrlOne))
	soundcloudUrlTwo := SoundcloudUrl{Url: "soundcloud.com/numbertwo"}
	err = soundcloudUrlService.AddSoundcloudUrl(soundcloudUrlTwo.Url)
	if err != nil {
		suite.T().Fail()
		return
	}
	soundcloudUrls, err = soundcloudUrlService.GetAllSoundcloudUrls()
	if err != nil {
		suite.T().Fail()
	}
	assert.True(suite.T(), soundcloudUrlExists(soundcloudUrls, soundcloudUrlTwo))
	newUiOrderOne := SoundcloudUrl{Url: soundcloudUrlOne.Url, UiOrder: 23}
	newUiOrderTwo := SoundcloudUrl{Url: soundcloudUrlTwo.Url, UiOrder: 5}
	err = soundcloudUrlService.UpdateSoundcloudUiOrders([]SoundcloudUrl{newUiOrderTwo, newUiOrderOne})
	if err != nil {
		suite.T().Fail()
	}
	soundcloudUrls, err = soundcloudUrlService.GetAllSoundcloudUrls()
	if err != nil {
		suite.T().Fail()
	}
	assert.True(suite.T(), soundcloudUrlExists(soundcloudUrls, newUiOrderOne))
	assert.True(suite.T(), soundcloudUrlExists(soundcloudUrls, newUiOrderTwo))
}

func soundcloudUrlExists(soundcloudUrls []SoundcloudUrl, url SoundcloudUrl) bool {
	for _, value := range soundcloudUrls {
		if value.Url == url.Url && value.UiOrder == url.UiOrder {
			return true
		}
	}
	return false
}
