/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package rcode

import (
	"fmt"
	"os"
)

// ReturnCode : return the R code that is used to run the analysis
func ReturnCode(outputDirectory string) {
	var codeString = `
#clkm1.1
library(MASS)


### Added by Paul to see if we can get input from command line
args <- commandArgs(trailingOnly = TRUE);

intensityfile = args[1];
thetafile = args[2];
analysis = args[3];
workingdir = args[4]
#workingdir = 'C:/Users/pauld/Desktop/Work Stuff'
#analysis = 'cl.km.1.1.abcd.exome'





kk = function(theta,intens,cut.12=55,cut.23=45,mindist=0.075,  
              cut.12.IN=65,cut.23.IN=65,cut.NuLL=0.09,
              cut.mean.NuLL=0.20,cut.range.NuLL = 0.50,cut.pcNuLLs=75,
              calling="germplasm",transformation="none",seed=13579) {

  # deal with missing values 
  
  i.ok.theta = which(!is.na(theta));  i.ok.intens = which(!is.na(intens))
  i.ok = intersect(i.ok.theta,i.ok.intens)
  THeta = theta[i.ok];  INtens = intens[i.ok]
  n.miss = length(theta) - length(THeta)
  n.geno = length(THeta)
  i.geno = c(1:length(THeta))
  
  
  n.NuLL=0
  i.NuLL = numeric(0)
  i.EXP = c(1:length(THeta))
  n.EXP = length(THeta)
  
  clstats = rep(0,times=32)
  clstats[c(4,5,6,8:19)] = NA
  CLids = rep(NA,times=n.geno)
  clids = rep(NA,times=length(theta))
  
  # Search Intens for NuLLs
  # Three features to identify; overall low response; overall wide range; individual low responses.
  # If wide range, cluster and assign the lowest cluster as NuLLs if its mean < cut.mean.NuLL
  # otherwise if low responder, leave alone
  # otherwise if there are some individual responses < cut.NuLL, assign these as NuLLs
  
  set.seed(seed)
  q.INtens = quantile(INtens,c(.05,.50,.95))
  low.responder = q.INtens[3]<cut.mean.NuLL
  wide.range = q.INtens[3] - q.INtens[1] > cut.range.NuLL
  clustering.INtens = 'FALSE'
  n.NuLL.clust = 0
  min.mean.IN = 0
  
  if (wide.range) { # perform full clustering and set the cluster with the lowest mean as NuLLs 
                    # if that mean is less than cut.mean.NuLL 
    clustering.INtens = 'TRUE'
   
    g1.IN = kmeans(INtens,1,nstart=5) 
    g2.IN = kmeans(INtens,2,nstart=5)  
    g3.IN = kmeans(INtens,3,nstart=5) 
    g4.IN = kmeans(INtens,4,nstart=5) 
    g5.IN = kmeans(INtens,5,nstart=5)
    list.g.IN = list(g1.IN,g2.IN,g3.IN,g4.IN,g5.IN)
    # grab reduction in within sums of squares results
    g2pc.IN = (g1.IN$tot.withinss-g2.IN$tot.withinss)/g1.IN$tot.withinss*100
    g3pc.IN = (g2.IN$tot.withinss-g3.IN$tot.withinss)/g2.IN$tot.withinss*100
    g4pc.IN = (g3.IN$tot.withinss-g4.IN$tot.withinss)/g3.IN$tot.withinss*100
    g5pc.IN = (g4.IN$tot.withinss-g5.IN$tot.withinss)/g4.IN$tot.withinss*100
    gpc.IN = round(c(g2pc.IN,g3pc.IN,g4pc.IN,g5pc.IN),digits=1)
    
    g1c.IN = g1.IN$centers
    g2c.IN = g2.IN$centers
    g3c.IN = g3.IN$centers
    g4c.IN = g4.IN$centers 
    g5c.IN = g5.IN$centers
    list.gc.IN = list(g1c.IN,g2c.IN,g3c.IN,g4c.IN,g5c.IN)
    
    n.IN.clust = 1
    if(g2pc.IN>cut.12.IN) n.IN.clust=2
    if ((n.IN.clust==2) & (g3pc.IN>cut.23.IN)) n.IN.clust=3
    if ((n.IN.clust==3) & (g4pc.IN>cut.23.IN)) n.IN.clust=4
    
    min.mean.IN = min(list.gc.IN[[n.IN.clust]])
    i.NuLL.clust = which( list.gc.IN[[n.IN.clust]]==min(list.gc.IN[[n.IN.clust]]) )#identify the NuLL cluster
    
    if (min.mean.IN<cut.mean.NuLL) { 
      i.NuLL = which(list.g.IN[[n.IN.clust]]$cluster==i.NuLL.clust) #identify its genotypes
      #i.NuLL.2 = which(list.g.IN[[n.IN.clust]]$cluster==i.NuLL.clust) #identify its genotypes
      #i.NuLL.3 = i.EXP[i.NuLL.2] #identify these genotype positions in i.EXP
      #i.NuLL = c(i.NuLL,i.NuLL.3)
      #print(i.NuLL.2)
      #print(i.NuLL.3)
      #print(i.NuLL)
      n.NuLL = length(i.NuLL)
      n.NuLL.clust = 1
      i.EXP =  setdiff(1:n.geno,i.NuLL)
    }
  }
  else {
   if  ( (min(INtens)<cut.NuLL) & !(low.responder) ) {
     #Find NuLLs
     i.NuLL = which(INtens<cut.NuLL)
     n.NuLL = length(i.NuLL)
     n.NULL.clust=1
     i.EXP = setdiff(1:n.geno,i.NuLL)}
}
 
    n.NuLL = length(i.NuLL)
    n.EXP = length(i.EXP)
    # sSmmary stats for Intensity data
    clstats[21] = round(q.INtens[3], digits=3)
    clstats[22] = round((q.INtens[3] - q.INtens[1]), digits=3)
    clstats[23:32] = 0
    if (clustering.INtens) { 
      locs4 = c(29:(29+n.IN.clust-1))
      clstats[23] = n.IN.clust
      clstats[24] = round(min.mean.IN, digits=2)
      clstats[25:28] = round(gpc.IN, digits=2) #% reduction in within sum of squares for INtens
      clstats[locs4] = round(c(as.numeric(list.gc.IN[[n.IN.clust]])), digits=2)} # INtensity cluster means
    
    
  # Abandon THeta clustering if too many low INtensities
  if (n.NuLL > (cut.pcNuLLs*n.geno/100)) {
   clstats[1] = n.geno
   clstats[3] = n.NuLL
   clstats[7] = round((n.NuLL/n.geno), digits=2) 
   clstats[20] = 1 # flag
   clids[i.NuLL] = ' N'
    } else 
    {
    # Remove any NuLLs from THeta 
  if (n.NuLL>0) THeta = THeta[i.EXP]
  # derive kmeans clusters using THeta for nclusters = 1:5
  set.seed(seed)
  
  if (transformation=='logit') THeta = logit(THeta)
  if (transformation=='ang')   THeta = asin(sqrt(THeta))
  
  g1 = kmeans(THeta,1,nstart=40) 
  g2 = kmeans(THeta,2,nstart=40)  
  g3 = kmeans(THeta,3,nstart=40) 
  g4 = kmeans(THeta,4,nstart=40) 
  g5 = kmeans(THeta,5,nstart=40)
  list.g = list(g1,g2,g3,g4,g5)
  # grab reduction in within sums of squares results
  g2pc = (g1$tot.withinss-g2$tot.withinss)/g1$tot.withinss*100
  g3pc = (g2$tot.withinss-g3$tot.withinss)/g2$tot.withinss*100
  g4pc = (g3$tot.withinss-g4$tot.withinss)/g3$tot.withinss*100
  g5pc = (g4$tot.withinss-g5$tot.withinss)/g4$tot.withinss*100
  gpc = round(c(g2pc,g3pc,g4pc,g5pc),digits=1)
  # grab cluster sizes
  g1n = g1$size
  g2n = g2$size
  g3n = g3$size
  g4n = g4$size
  g5n = g5$size
  list.n = list(g1n,g2n,g3n,g4n,g5n)
  #  derive cluster sds
  g1sd = sqrt(g1$withinss/(ifelse(g1n>1,g1n-1,1)))
  g2sd = sqrt(g2$withinss/(ifelse(g2n>1,g2n-1,1)))
  g3sd = sqrt(g3$withinss/(ifelse(g3n>1,g3n-1,1)))
  g4sd = sqrt(g4$withinss/(ifelse(g4n>1,g4n-1,1)))
  g5sd = sqrt(g5$withinss/(ifelse(g5n>1,g5n-1,1)))
  #print(g2$withinss[1:2]))
  #print(g2sd)
  list.gsd = list(g1sd,g2sd,g3sd,g4sd,g5sd)
  # grab cluster means
  g1c = g1$centers
  g2c = g2$centers
  g3c = g3$centers
  g4c = g4$centers 
  g5c = g5$centers
  list.gc = list(g1c,g2c,g3c,g4c,g5c)
  
  
  # derive numbere of clusters and fill the arguments (clstats and clids)
  n.clust = 1
  if(g2pc>cut.12) { # subject to mindist constraint we have at least 2 clusters.  Are there 3?
      d12 = abs(g3c[1]-g3c[2])
      d13 = abs(g3c[1]-g3c[3])
      d23 = abs(g3c[2]-g3c[3])
      if((g3pc>cut.23) & (min(c(d12,d13,d23))>mindist)) n.clust=3 
      else {
        d12 = abs(g2c[1]-g2c[2])
        if (d12>mindist) n.clust=2}
    }
  
  n.clustNuLL = n.clust+(n.NuLL>0)
  clstats[1] = n.geno
  clstats[2] = n.clust
  clstats[3] = n.NuLL
  clstats[18] = 0
  clstats[19] = 0
  # define some locations for clstats
  locs1 = c(4:(4+(n.clust-1)))
  locs2 = c(8:(8+(n.clust-1)))
  locs3 = c(11:(11+(n.clust-1)))
  # sort into ascending order the mean theta values of the clusters
  ord = order(as.numeric(list.gc[[n.clust]]))
  clstats[locs1] = round(c(list.n[[n.clust]][ord]/(n.geno-n.NuLL)), digits=2) #proportion in each cluster
  clstats[7] = round((n.NuLL/n.geno)*(n.NuLL>0), digits=2)
  clstats[locs2] = round(c(as.numeric(list.gc[[n.clust]][ord])), digits=3) # mean theta values
  clstats[locs3] = round(c(as.numeric(list.gsd[[n.clust]][ord])), digits=3) #sd theta values
  clstats[14:17] = round(gpc, digits=2) #% reduction in within sum of squares
  
  if (n.clust>1) {
    sdmn = sd(as.numeric(list.gc[[n.clust]])) #Separation
    if (n.clust==2) clstats[18] = sdmn/0.71 # scale to[0,1]
    if (n.clust==3) clstats[18] = sdmn/0.55 } # scale to [0,1]
  clstats[18] = round(clstats[18], digits=2)
  lmnsd = log(mean(as.numeric(list.gsd[[n.clust]]), na.rm=T)) # Compactess
  clstats[19] = round((-1.5-lmnsd)/5.5, digits=2) #scale to [1,0]; range is (-7.0 to -1.5)
  # set flag for large ThetA %reductions in totwithinss and low (q.025) Sep or Comp 
  clstats[20] = as.numeric( (g4pc>cut.12) | ( (clstats[18]<0.12) & (n.clust>1) )| clstats[19]<0.25 )  
  # Add summary stats for Intensity data
  #clstats[21] = round(q.INtens[3], digits=3)
  #clstats[22] = round((q.INtens[3] - q.INtens[1]), digits=3)
  #clstats[23:32] = 0
  #if (clustering.INtens) { 
  #  locs4 = c(29:(29+n.IN.clust-1))
  #  clstats[23] = n.IN.clust
  #  clstats[24] = round(min.mean.IN, digits=2)
  #  clstats[25:28] = round(gpc.IN, digits=2) #% reduction in within sum of squares for INtens
  #  clstats[locs4] = round(c(as.numeric(list.gc.IN[[n.IN.clust]])), digits=2)} # INtensity cluster means
  
  if (calling=="germplasm") {
    CLlabels1 = c('AA')
    if ((n.clust==1) & mean(THeta)>0.5) CLlabels1 = c('BB')
    CLlabels2 = c('AA','BB')
    CLlabels3 = c('AA','AB','BB') } else {
      # mapping pop
    CLlabels1 = c('AA')
    if (n.clust==1) { if (mean(THeta)<0.35) CLlabels1=c('AA')  else {
      if (mean(THeta)<0.65) CLlabels1=c('AB') else CLlabels1 = c('BB')}} 
    CLlabels2=c("AA","BB")
    if (g2c[1]>0.35) CLlabels2 = c('AB','BB')
    if (g2c[2]<0.65) CLlabels2 = c('AA','AB')
    CLlabels3 = c('AA','AB','BB')
    }
  list.CLlabels = list(CLlabels1,CLlabels2,CLlabels3)
  tempids = factor(list.g[[n.clust]]$cluster, levels=ord, labels=list.CLlabels[[n.clust]])
  CLids[i.EXP] = as.character(tempids) # place the ids in the EXpressed genotypes matrix
  
  if (n.NuLL>0) CLids[i.NuLL] = ' N' # add the NuLL ids
  
  clids[i.ok.theta] = CLids} # place above in complete matrix 
  
  return(list(stats=clstats,ids=clids))
}


########################################################################################################

 
###
### Set the clustering arguments
###
 
my.cut.12 = 55
my.cut.23 = 45
my.cut.12.IN = 65
my.cut.23.IN = 65
my.mindist=0.075
my.cut.NuLL = 0.09
my.cut.mean.NuLL=0.20
my.cut.range.NuLL = 0.50
my.cut.pcNuLLs=75
my.calling="germplasm"
my.seed=13579
my.transformation = "none"
 

###
### Read all data and grab n.geno & n.snp
###

setwd(workingdir)

alltheta=read.csv(thetafile,h=T,stringsAsFactors=F)
n.snp = length(alltheta[,1])
n.geno = ncol(alltheta)-1
n.geno1 = n.geno+1

names.snp = alltheta[,1]
line1alltheta=read.csv(thetafile,h=F,stringsAsFactors=F,nrows=1)
names.geno = line1alltheta[2:n.geno1]

tall.theta = t(alltheta[,2:ncol(alltheta)])
df.tall.theta = data.frame(tall.theta); colnames(df.tall.theta) = names.geno

allintens=read.csv(intensityfile,h=T,stringsAsFactors=F)
tall.intens = t(allintens[,2:ncol(allintens)])
df.tall.intens = data.frame(tall.intens); colnames(df.tall.intens) = names.geno
# remove non-dataframe copies of the data
rm("alltheta", "tall.theta", "allintens", "tall.intens")

file.stats = paste(analysis,"stats","csv", sep=".")
file.ids =   paste(analysis,"ids",  "csv", sep=".")
file.parms = paste(analysis,"parms","csv", sep=".")

parm.values=c()
parm.names = c('Clustering','Date & Time','workingdir', 'file.stats', 'file.ids', 'intensityfile', 'thetafile',
               'cut.12.IN', 'cut.23.IN','mindist','minsize', 'range.adj', 
               'cut.NuLL', 'cut.mean.NuLL', 'cut.range.NuLL', 'cut.pcNuLLs', 'seed',
               'calling',  'transformation','blank','blank','blank')

date.time = Sys.time()
parm.values[1] = 'K Means'
parm.values[2] = as.character(date.time)
parm.values[3:7] =c(workingdir, file.stats, file.ids, intensityfile, thetafile)
parm.values[8:17] = as.character(c(my.cut.12.IN, my.cut.23.IN,my.mindist,NA,NA, 
                                   my.cut.NuLL, my.cut.mean.NuLL, my.cut.range.NuLL, my.cut.pcNuLLs,my.seed))
parm.values[18:22] = c(my.calling, my.transformation,'*','*','*' )

df.parm = data.frame( parm.names,parm.values)




###
###
### run a single SNP to check for faults
###
###

choice = 23208
kk(df.tall.theta[,choice],df.tall.intens[,choice],cut.12=my.cut.12,cut.23=my.cut.23,mindist=my.mindist,
   cut.12.IN=my.cut.12.IN,cut.23.IN=my.cut.NuLL,cut.NuLL=my.cut.NuLL,cut.mean.NuLL=my.cut.mean.NuLL,
   cut.range.NuLL=my.cut.range.NuLL,cut.pcNuLLs=my.cut.pcNuLLs,
   calling=my.calling,transformation=my.transformation,seed=my.seed) 


###
###
###
### run for all SNPs
###
###

temp =mapply(kk,df.tall.theta[,1:n.snp], df.tall.intens[,1:n.snp],cut.12=my.cut.12,cut.23=my.cut.23,mindist=my.mindist,
            cut.12.IN=my.cut.12.IN,cut.23.IN=my.cut.NuLL,cut.NuLL=my.cut.NuLL,cut.mean.NuLL=my.cut.mean.NuLL,
          cut.range.NuLL=my.cut.range.NuLL,cut.pcNuLLs=my.cut.pcNuLLs,
          calling=my.calling,transformation=my.transformation,seed=my.seed)
 

 
names.stats =c("ngeno","nclusters","nNulls","p1","p2","p3","pNULL","mn1","mn2","mn3",
                              "sd1","sd2","sd3","%1/2","%2/3","%3/4","%4/5","Sep","Comp","flag",
                              "INq95","INRange","n.IN.cl","IN.MinCl","IN%1/2","IN%2/3","IN%3/4","IN%4/5",
                              "INmn1","INmn2","IN.mn3","IN.mn4")

odds = c(1+2*0:(n.snp-1)) 
stats = matrix(data=unlist(temp[odds]), nrow=n.snp, ncol=32, byrow=T) 
df.stats = as.data.frame(stats, row.names = names.snp)
names(df.stats) = names.stats  
df.stats[,33] = c(1:n.snp)
colnames(df.stats)[33] = "orig.snp.id"

#Grab the cluster allocations (even columns of the results list) and put into a dataframe
evens = c(2*(1:n.snp))
ids = matrix(data=unlist(temp[evens]), nrow=n.snp, ncol=n.geno, byrow=T)
df.ids = as.data.frame(ids, row.names=names.snp) 
names(df.ids) = names.geno
df.ids[,(n.geno+1)] = c(1:n.snp)
colnames(df.ids)[(n.geno+1)] = "orig.snp.id"

options(digits=3)
write.csv(df.parm,file.parms)
write.csv(df.stats, na='*', file.stats)
write.csv(df.ids, file.ids)

 
#a 16/06/18  changged kk to deal with different missing patterns in Theta & Intens files
#b 17/06/18  added clling argument to kk
#c 20/06/18  added 'parms.csv' to output
#d 22/06/18  changed way theta file is read, to preserve genotype names exactly
`

	// Now print the file out locally this can be kept as a record of the code that was run
	outFile, err := os.Create(outputDirectory + "/jim_modified.R")
	if err != nil {
		fmt.Println(err)
	}
	outFile.WriteString(codeString)

}
