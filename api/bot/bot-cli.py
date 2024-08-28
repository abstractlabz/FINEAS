import discord
from discord.ext import commands
import requests
import os
from dotenv import load_dotenv

load_dotenv()

DISCORD_TOKEN = os.getenv('DISCORD_TOKEN')
RETRIEVAL_API_URL = os.getenv('RETRIEVAL_API_URL')
CHATBOT_API_URL = os.getenv('CHATBOT_API_URL')

bot = commands.Bot(command_prefix='!', intents=discord.Intents.all())

@bot.event
async def on_ready():
    print(f'{bot.user} has connected to Discord!')

@bot.command(name='ticker', help='Get information about a specific ticker')
async def ticker(ctx, symbol: str):
    try:
        response = requests.get(f"{RETRIEVAL_API_URL}/ret?ticker={symbol}")
        if response.status_code == 200:
            data = response.json()
            embed = discord.Embed(title=f"Information for {symbol}", color=discord.Color.blue())
            embed.add_field(name="Stock Performance", value=data.get('StockPerformance', 'N/A'), inline=False)
            embed.add_field(name="Financial Health", value=data.get('FinancialHealth', 'N/A'), inline=False)
            embed.add_field(name="News Summary", value=data.get('NewsSummary', 'N/A'), inline=False)
            embed.add_field(name="Company Description", value=data.get('CompanyDesc', 'N/A'), inline=False)
            embed.add_field(name="Technical Analysis", value=data.get('TechnicalAnalysis', 'N/A'), inline=False)
            await ctx.send(embed=embed)
        else:
            await ctx.send(f"Error retrieving information for {symbol}. Please try again later.")
    except Exception as e:
        await ctx.send(f"An error occurred: {str(e)}")

@bot.command(name='ask', help='Ask an investment research question')
async def ask(ctx, *, question: str):
    try:
        response = requests.post(f"{CHATBOT_API_URL}/chat", json={"prompt": question})
        if response.status_code == 200:
            answer = response.text
            await ctx.send(f"Answer: {answer}")
        else:
            await ctx.send("Error processing your question. Please try again later.")
    except Exception as e:
        await ctx.send(f"An error occurred: {str(e)}")

bot.run(DISCORD_TOKEN)
