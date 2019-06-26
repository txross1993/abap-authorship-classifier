# Purpose
The purpose of this project is to use supervised machine learning techniques to determine features that are most deterministic of authorship of abap code.

Code authorship identification is not a new problem domain in machine learning. Techniques and experiments have been designed to extract relevant style and structure features of Java, C-languages, and Python code. [This article](http://injoit.org/index.php/j1/article/view/665/655) provides an overview of the problem solving methodology applied in various experiments over the years. 

From the article:
>> Many software products that solve problems associated
with the style of writing code are based on the use of various
methods of machine learning. Traditional methodology for
obtaining software for the task, used in this area, usually
involves the following steps:
>>1. Extracting software metrics that could define an
author’s style
>>2. Filtering metrics and highlighting the really
significant ones
>>3. Choosing a machine learning model for classifying
and training the model using selected metrics
>>4. The application of the model is based on the
selection of an already filtered set of metrics.

<p style="text-align:left"><i>(p.3 of pdf)</i></p>

# Methodology

Methods involve extracting features from input "text" (i.e. the code).

This project attempts to explore feature extraction on:
  1. Embeddings on raw abap code
  2. Concrete Syntax Trees derived from a chevrotain-based parser

## Embeddings on Raw Text

I will be mimicking logic described in [this article](https://towardsdatascience.com/building-machine-learning-model-from-unstructured-data-dd2d0263f1db), which uses TF-IDF embeddings on text lists of ingredients to determine cuisine. 

An alternative to TF-IDF is Word2Vec. Both of these tools provide ways to structure raw text into vectorized inputs to a data model. These methods can be used to determine relationships between and among words. Hypothetically, I could use the vectorized relationships as input features to a model to classify a program file's author.

 - [This article](http://blog.christianperone.com/2011/09/machine-learning-text-feature-extraction-tf-idf-part-i/) provides an introductory tutorial to TF-IDF
 - [And this](https://skymind.ai/wiki/word2vec) describes Word2Vec

## Concrete Syntax Trees

In an attempt to parse raw text into syntax trees, I am exploring a tool called [chevrotain](https://github.com/SAP/chevrotain), a javascript-based parsing toolkit. The thought behind exploring this avenue as opposed to the word embeddings is that while ABAP is very English-like in grammar, it is a programming language. Syntax trees can abstract the structural elements of an ABAP program, which can be collected, measured, and provided as input features to a classifier network. 

Initial work can be found in the branch [chevrotain-tokens](https://github.com/txross1993/abap-authorship-classifier/tree/chevrotain-tokens)
