import React, {Component} from 'react';
import ReactPlayer from "react-player"

export class AudioPlayer extends Component<any, any> {
    render() {
        return (
            <div className="audioPlayer">
                <ReactPlayer
                    url = {this.props.soundcloudUrl + "?show_teaser=false"}
                    className='react-player'
                    config={{
                        soundcloud: {
                            // should work according to docs https://github.com/CookPete/react-player but it's borked. I shove the data in url as query param lmao ??
                            options : {
                                show_user: false,
                                style:"text-decoration: none;"
                            }
                        }
                    }}
                />
            </div>
        );
    }
}
