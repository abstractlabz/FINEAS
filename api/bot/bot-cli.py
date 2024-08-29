import discord
from discord.ext import commands
import requests
import os
import urllib.parse
from dotenv import load_dotenv

load_dotenv()

DISCORD_TOKEN = os.getenv('DISCORD_TOKEN') or ""
RETRIEVAL_API_URL = os.getenv('RETRIEVAL_API_URL')
CHATBOT_API_URL = os.getenv('CHATBOT_API_URL')
AUTH_BEARER = os.getenv('AUTH_BEARER')

bot = commands.Bot(command_prefix='!', intents=discord.Intents.all())

@bot.event
async def on_ready():
    print(f'{bot.user} has connected to Discord!')

@bot.command(name='ticker', help='Get information about a specific ticker')
async def ticker(ctx: commands.Context, symbol: str) -> None:
    try:
        headers = {"Authorization": f"Bearer {AUTH_BEARER}"}
        response = requests.get(f"{RETRIEVAL_API_URL}/ret?ticker={symbol}", headers=headers)
        print(f"Response Status Code: {response.status_code}")
        print(f"Response Content: {response.content}")
        data = response.json()
        embed = discord.Embed(title=f"Information for {symbol}", color=discord.Color.blue())
        embed.add_field(name="Stock Performance", value=data.get('StockPerformance', 'N/A'), inline=False)
        await ctx.send(embed=embed)
    except Exception as e:
        await ctx.send(f"Error retrieving information for {symbol}. Please try again later.")
        await ctx.send(f"An error occurred: {str(e)}")

@bot.command(name='ask', help='Ask an investment research question')
async def ask(ctx: commands.Context, *, question: str) -> None:
    try:
        question = urllib.parse.quote(question)
        headers = {"Authorization": f"Bearer {AUTH_BEARER}"}
        response = requests.post(f"{CHATBOT_API_URL}/chat", json={"prompt": question}, headers=headers)
        answer = response.text
        await ctx.send(f"Answer: {answer}")
    except Exception as e:
        await ctx.send("Error processing your question. Please try again later.")
        await ctx.send(f"An error occurred: {str(e)}")

if DISCORD_TOKEN:
    bot.run(DISCORD_TOKEN)
else:
    print("Error: DISCORD_TOKEN is not set.")