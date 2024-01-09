import {
    APIInteractionResponse,
    InteractionResponseType,
} from "discord-api-types/v10";
import { AutocompleteHandler, CommandHandler, getOption } from "../common.ts";

const USER_AGENT = `ExampleBunBot/0.2.2 (https://github.com/sportshead/powergrid; powergrid@sportshead.dev) bun/${Bun.version}`;

const getWikiPage = async (
    title: string,
): Promise<{
    title: string;
    content_urls: {
        desktop: {
            page: string;
        };
    };
    extract: string;
    timestamp: string;
    thumbnail: {
        source: string;
    };
    description: string;
}> => {
    let res: Response;
    do {
        console.log(
            `[${new Date().toUTCString()}] Getting wiki page "${title}"`,
        );
        res = await fetch(
            `https://en.wikipedia.org/api/rest_v1/page/summary/${title}`,
            {
                headers: {
                    "User-Agent": USER_AGENT
                }
            }
        );
        title = res.headers.get("Location") ?? "";
    } while (res.status === 301 || res.status === 302);
    return res.json();
};
export const wikiCommandHandler: CommandHandler = async (interaction) => {
    const title = getOption<string>(interaction, "title", "Earth");
    const page = await getWikiPage(title);

    const res: APIInteractionResponse = {
        type: InteractionResponseType.ChannelMessageWithSource,
        data: {
            embeds: [
                {
                    title: page.title,
                    url: page.content_urls.desktop.page,
                    timestamp: new Date(page.timestamp).toISOString(),
                    description: page.extract,
                    thumbnail: {
                        url: page.thumbnail.source,
                    },
                    footer: {
                        text: page.description,
                    },
                    author: {
                        name: "Wikipedia",
                        url: "https://en.wikipedia.org",
                        icon_url:
                            "https://upload.wikimedia.org/wikipedia/commons/thumb/8/80/Wikipedia-logo-v2.svg/526px-Wikipedia-logo-v2.svg.png",
                    },
                },
            ],
        },
    };

    return new Response(JSON.stringify(res), {
        status: 200,
        headers: { "Content-Type": "application/json" },
    });
};
export const wikiAutocompleteHandler: AutocompleteHandler = async (
    interaction,
) => {
    const query = getOption<string>(interaction, "title", "");
    if (!query) {
        return new Response(null, { status: 400 });
    }
    const search: {
        pages: {
            id: number;
            key: string;
            title: string;
            excerpt: string;
            description: string;
        }[];
    } = await fetch(
        `https://api.wikimedia.org/core/v1/wikipedia/en/search/page?q=${encodeURIComponent(
            query,
        )}&limit=25`,
        {
            headers: {
                "User-Agent": USER_AGENT
            }
        }
    ).then((r) => r.json());

    const res: APIInteractionResponse = {
        type: InteractionResponseType.ApplicationCommandAutocompleteResult,
        data: {
            choices: search.pages.map((page) => ({
                name: `${page.title} - ${page.description}`,
                value: page.title,
            })),
        },
    };

    return new Response(JSON.stringify(res), {
        status: 200,
        headers: { "Content-Type": "application/json" },
    });
};
