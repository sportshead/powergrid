import {
    APIInteractionResponse,
    APIInteractionResponseCallbackData,
    ButtonStyle,
    ComponentType,
    InteractionResponseType,
    MessageFlags,
    TextInputStyle,
} from "discord-api-types/v10";
import {
    BUN_INTERACTION_PREFIX,
    CommandHandler,
    ComponentHandler,
    getOption,
    json,
    ModalHandler,
} from "../common.ts";

const INTERACTION_PREFIX = `${BUN_INTERACTION_PREFIX}/counter`;
const createMessage = (
    name: string,
    count: number,
    initial: number,
): APIInteractionResponseCallbackData => ({
    allowed_mentions: {
        parse: [],
    },
    content: `**${name}**: ${count}`,
    components: [
        {
            type: ComponentType.ActionRow,
            components: [
                {
                    type: ComponentType.Button,
                    emoji: {
                        name: "➕",
                    },
                    custom_id: `${INTERACTION_PREFIX}/${name};${
                        count + 1
                    };${initial};inc`,
                    style: ButtonStyle.Primary,
                },
                {
                    type: ComponentType.Button,
                    emoji: {
                        name: "➖",
                    },
                    custom_id: `${INTERACTION_PREFIX}/${name};${
                        count - 1
                    };${initial};dec`,
                    style: ButtonStyle.Primary,
                },
                {
                    type: ComponentType.Button,
                    emoji: {
                        name: "🔄",
                    },
                    custom_id: `${INTERACTION_PREFIX}/${name};${initial};${initial};res`,
                    style: ButtonStyle.Secondary,
                },
                {
                    type: ComponentType.Button,
                    emoji: {
                        name: "✏️",
                    },
                    custom_id: `${INTERACTION_PREFIX}/${name};${count};${initial}/edit`,
                    style: ButtonStyle.Secondary,
                },
                {
                    type: ComponentType.Button,
                    emoji: {
                        name: "🗑️",
                    },
                    custom_id: `${INTERACTION_PREFIX}/${name};${count};${initial}/delete`,
                    style: ButtonStyle.Danger,
                },
            ],
        },
    ],
});

export const counterCommandHandler: CommandHandler = async (interaction) => {
    const name = getOption<string>(interaction, "name", "Counter");
    const initial = getOption<number>(interaction, "value", 0);

    if (name.includes(";") || name.includes("/")) {
        const res: APIInteractionResponse = {
            type: InteractionResponseType.ChannelMessageWithSource,
            data: {
                content: "Counter name cannot contain `;` or `/`",
                flags: MessageFlags.Ephemeral,
            },
        };
        return json(res);
    }

    const res: APIInteractionResponse = {
        type: InteractionResponseType.ChannelMessageWithSource,
        data: createMessage(name, initial, initial),
    };

    console.log(
        `[${new Date().toUTCString()}] Responding to interaction ${
            interaction.id
        } from @${interaction.member?.user?.username} (${interaction.member
            ?.user?.id}`,
        res,
    );

    return json(res);
};

export const counterComponentHandler: ComponentHandler = async (
    interaction,
) => {
    const [params, action] = interaction.data.custom_id.split("/").slice(2);
    const [name, _count, _initial] = params.split(";");
    const count = parseInt(_count);
    const initial = parseInt(_initial);
    if (action === "edit") {
        const res: APIInteractionResponse = {
            type: InteractionResponseType.Modal,
            data: {
                title: `Edit "${name}"`,
                custom_id: `${INTERACTION_PREFIX}/${name};${count};${initial}`,
                components: [
                    {
                        type: ComponentType.ActionRow,
                        components: [
                            {
                                type: ComponentType.TextInput,
                                custom_id: "name",
                                label: "Name",
                                style: TextInputStyle.Short,
                                value: name,
                                required: true,
                                max_length: 32,
                            },
                        ],
                    },
                    {
                        type: ComponentType.ActionRow,
                        components: [
                            {
                                type: ComponentType.TextInput,
                                custom_id: "value",
                                label: "Value",
                                style: TextInputStyle.Short,
                                required: true,
                                value: count.toString(),
                            },
                        ],
                    },
                    {
                        type: ComponentType.ActionRow,
                        components: [
                            {
                                type: ComponentType.TextInput,
                                custom_id: "initial",
                                label: "Initial Value",
                                style: TextInputStyle.Short,
                                required: true,
                                value: initial.toString(),
                            },
                        ],
                    },
                ],
            },
        };
        return json(res);
    }
    if (action === "delete") {
        fetch(
            `https://discord.com/api/webhooks/${interaction.application_id}/${interaction.token}/messages/@original`,
            {
                method: "DELETE",
            },
        ).then(
            (r): Promise<any> =>
                r.status !== 204
                    ? r
                          .json()
                          .then((body) =>
                              console.error(
                                  `[${new Date().toUTCString()}] Failed to delete message`,
                                  r.status,
                                  r.statusText,
                                  body,
                              ),
                          )
                    : fetch(
                          `https://discord.com/api/webhooks/${interaction.application_id}/${interaction.token}`,
                          {
                              method: "POST",
                              headers: {
                                  "Content-Type": "application/json",
                              },
                              body: JSON.stringify({
                                  content: `Deleted counter **${name}** with value ${count}`,
                                  flags: MessageFlags.Ephemeral,
                              }),
                          },
                      ),
        );

        const res: APIInteractionResponse = {
            type: InteractionResponseType.DeferredMessageUpdate,
        };
        return json(res);
    }
    const res: APIInteractionResponse = {
        type: InteractionResponseType.UpdateMessage,
        data: createMessage(name, count, initial),
    };

    console.log(
        `[${new Date().toUTCString()}] Responding to interaction ${
            interaction.id
        } ${interaction.token} from @${interaction.member?.user
            ?.username} (${interaction.member?.user?.id}`,
        res,
    );

    return json(res);
};

export const counterModalHandler: ModalHandler = async (interaction) => {
    const [name, _count, _initial] = interaction.data.custom_id
        .split("/")[2]
        .split(";");
    const count = parseInt(_count);
    const initial = parseInt(_initial);

    const data = {
        name,
        count,
        initial,
    };
    for (const component of interaction.data.components.flatMap(c=>c.components)) {
        if (component.type !== ComponentType.TextInput) {
            continue;
        }
        if (component.custom_id === "name") {
            data.name = component.value;
        }
        if (component.custom_id === "value") {
            const int = parseInt(component.value);
            data.count = isNaN(int) ? count : int;
        }
        if (component.custom_id === "initial") {
            const int = parseInt(component.value);
            data.initial = isNaN(int) ? initial : int;
        }
    }

    const res: APIInteractionResponse = {
        type: InteractionResponseType.UpdateMessage,
        data: createMessage(data.name, data.count, data.initial),
    };

    console.log(
        `[${new Date().toUTCString()}] Responding to interaction ${
            interaction.id
        } from @${interaction.member?.user?.username} (${interaction.member
            ?.user?.id}`,
        res,
    );

    return json(res);
};
