import { BearerToken } from './andrewwillette'

// test('Test getSoundcloudUrls returns sally ann', async () => {
// 	const soundcloudUrls: Promise<HttpResponse<SoundcloudUrl[]>> = getSoundcloudUrls()
// 	const result_1 = await soundcloudUrls;
// 	const expectedSong = { url: 'https://soundcloud.com/user-434601011/sally-ann' };
// 	expect(result_1.parsedBody).toContain(expectedSong);
// });

test('BearerToken properties', () => {
	const token: BearerToken = { bearerToken: "testToken" }
	expect(token.bearerToken).toEqual("testToken")
});
